package mysql

import (
	"net"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mysql/protocol"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/ssl"
)

/*
AuthenticationHandshake represents the entire back and forth process
between a MySQL client and server during which authentication occurs.
Note this is distinct from the various specific handshake packets that are
sent back and forth during this process.

Note this is process is referred to as the "connection phase" in the MySQL
docs:

https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase.html

Overview of the Handshake Process

	+---------+                   +-------------+                         +---------+
	| Client  |                   | Secretless  |                         | Backend |
	+---------+                   +-------------+                         +---------+
		 |                               |                                     |
		 |                               |              InitialHandshakePacket |
		 |                               |<------------------------------------|
		 |                               |                                     |
		 |        InitialHandshakePacket |                                     |
		 |<------------------------------|                                     |
		 |                               |                                     |
		 | HandshakeResponse             |                                     |
		 |------------------------------>|                                     |
		 |                               | -------------------------------\    |
		 |                               |-| If client requested SSL, add |    |
		 |                               | | it to HandshakeResponse      |    |
		 |                               | |------------------------------|    |
		 |                               | HandshakeResponse                   |
		 |                               |------------------------------------>|
		 |                               |                                     |
		 |                               |                            OkPacket |
		 |                               |<------------------------------------|
		 |                               |                                     |
		 |                      OkPacket |                                     |
		 |<------------------------------|                                     |
		 |                               |                                     |

Note: The above diagram was created using https://textart.io/sequence and the
following source:

	object Client Secretless Backend
	Backend->Secretless: InitialHandshakePacket
	Secretless->Client: InitialHandshakePacket
	Client->Secretless: HandshakeResponse
	note right of Secretless: If client requested SSL, add\n it to HandshakeResponse
	Secretless->Backend: HandshakeResponse
	Backend->Secretless: OkPacket
	Secretless->Client: OkPacket
*/
type AuthenticationHandshake struct {
	connectionDetails *ConnectionDetails
	clientConn        *Connection
	backendConn       *Connection

	sslMode *ssl.DbSSLMode

	//TODO: after kumbi's work these 2 should be combined
	rawServerHandshake      Packet
	serverHandshake         *protocol.HandshakeV10
	clientHandshakeResponse *protocol.HandshakeResponse41

	err error
}

// NewAuthenticationHandshake creates a new AuthenticationHandshake command object,
// intended to be Run().
func NewAuthenticationHandshake(
	clientConn *Connection,
	backendConn *Connection,
	connDetails *ConnectionDetails,
) AuthenticationHandshake {
	return AuthenticationHandshake{
		connectionDetails: connDetails,
		clientConn:        clientConn,
		backendConn:       backendConn,
	}
}

// Run executes all the logic needed to complete authentication between a
// MySQL server and client.  When it completes successfully,
// AuthenticatedBackendConn will return the raw, authenticated network conn.
func (h *AuthenticationHandshake) Run() error {
	// The server is the first to communicate. Read the server handshake
	h.readServerHandshake()
	// Pass along the server handshake to the client, with some minor modifications
	//
	// 1. Remove TLS capability to avoid TLS connections to Secretless.
	// 2. Use `mysql_native_password` as the auth plugin between the client and Secretless, to make
	// life easier. We actually don't care about the credentials from the client, we just need the rest
	// of the packet.
	h.writeHandshakeToClient()

	// Read the client handshake response.
	//
	// We are done listening to the client!
	h.readClientHandshakeResponse()

	// Make sure if the connector (not the client) is configured to use TLS
	// then the server supports TLS. This must be done after reading the client response otherwise if validation
	// fails then the client connection hangs
	h.validateServerSSL()

	// Everything beyond this point is in service of responding to the server authentication challenge.

	// Override client capabilities. For example, the connector has secure connection capabilities and supports
	// authentication plugins.
	h.overrideClientCapabilities()
	// Inject credentials into the client handshake response
	h.injectCredentials()

	h.handleClientSSLRequest()

	// Write modified client handshake to server, and
	// carry out the rest of the authentication dance between Secretless and the server.
	// When we're done we just let the client know of the outcome.
	h.writeClientHandshakeResponseToBackend()
	h.handleBackendAuthResponse()

	return h.err
}

// AuthenticatedBackendConn returns an already authenticated connection
// to the MySQL server.  Intended to be called after Run() has completed.
func (h *AuthenticationHandshake) AuthenticatedBackendConn() net.Conn {
	return h.backendConn.RawConnection()
}

func (h *AuthenticationHandshake) readServerHandshake() {
	if h.err != nil {
		return
	}
	h.rawServerHandshake = h.readBackendPacket()

	if h.err != nil {
		return
	}
	h.serverHandshake, h.err = protocol.UnpackHandshakeV10(h.rawServerHandshake)
}

func (h *AuthenticationHandshake) writeHandshakeToClient() {
	if h.err != nil {
		return
	}

	serverHandshake := *h.serverHandshake
	// Remove Client SSL Capability from Server Handshake Packet
	// to force client to connect to Secretless without SSL
	// TODO: update this after kumbi's work
	serverHandshake.ServerCapabilities &^= protocol.ClientSSL

	// Give client the simplest auth plugin request
	// This might work for now, but we'll likely need to add support for other auth plugins
	serverHandshake.AuthPlugin = "mysql_native_password"

	packetWithNoSSL, err := protocol.PackHandshakeV10(&serverHandshake)
	if err != nil {
		h.err = err
		return
	}

	// TODO: push all packing code into the `WriteXXX` methods
	h.writeClientPacket(packetWithNoSSL)
}

func (h *AuthenticationHandshake) validateServerSSL() {
	if h.err != nil {
		return
	}

	if h.clientRequestedSSL() && !h.serverSupportsSSL() {
		h.err = protocol.ErrNoTLS
	}
}

func (h *AuthenticationHandshake) readClientHandshakeResponse() {
	if h.err != nil {
		return
	}

	rawResponse := h.readClientPacket()
	if h.err != nil {
		return
	}

	// TODO: client requesting SSL results in ERROR 2026 (HY000): SSL connection error: protocol version mismatch
	h.clientHandshakeResponse, h.err = protocol.UnpackHandshakeResponse41(rawResponse)
}

func (h *AuthenticationHandshake) overrideClientCapabilities() {
	if h.err != nil {
		return
	}

	// TODO: after kumbi's done, change below to method calls

	// TODO: add tests cases for authentication plugins support
	// Disable CapabilityFlag for authentication plugins support
	h.clientHandshakeResponse.CapabilityFlags |= protocol.ClientPluginAuth

	// TODO: add tests cases for client secure connection
	// Enable CapabilityFlag for client secure connection
	// TODO: explore weird heisenbug when this is toggled off:  ERROR: 1043 (08S01): Bad handshake
	h.clientHandshakeResponse.CapabilityFlags |= protocol.ClientSecureConnection

	// Ensure CapabilityFlag is set when using TLS
	if h.clientRequestedSSL() {
		h.clientHandshakeResponse.CapabilityFlags |= protocol.ClientSSL
	}

}

func (h *AuthenticationHandshake) injectCredentials() {
	if h.err != nil {
		return
	}

	// TODO: change this to method call on clientHandshakeResponse when Kumbi's work done
	h.err = protocol.InjectCredentials(
		h.serverHandshake.AuthPlugin,
		h.clientHandshakeResponse,
		h.serverHandshake.Salt,
		h.connectionDetails.Username,
		h.connectionDetails.Password,
	)
}

func (h *AuthenticationHandshake) handleClientSSLRequest() {
	if h.err != nil {
		return
	}

	if !h.clientRequestedSSL() {
		return
	}

	// The SSLRequest packet is created by copying the HandshakeResponse,
	// but truncating the username and everything after the username in
	// the payload, as described here:
	//
	// https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase_packets_protocol_ssl_request.html
	//
	// The payload itself breaks down as follows:
	//
	//  +------------+-----------------------------------------+
	//  | Num Bytes  |               Description               |
	//	+------------+-----------------------------------------+
	//  | 4          | capability flags, CLIENT_SSL always set |
	//  | 4          | max-packet size                         |
	//  | 1          | character set                           |
	//  | string[23] | reserved (all [0])                      |
	//	+------------+-----------------------------------------+
	//
	//  Hence by taking the first (4+4+1+23) bytes we take everything in
	//  the payload up to, but not including, the username.  The final
	//  +4 in (4+4+1+23)+4 accounts for the header section before the
	//  payload, ie, the payload_length and the sequence_id, as described
	//  in the comment above this one.

	// TODO: Note currently just repeating this logic. Will change after kumbi integration
	packedHandshakeRespPacket, err := protocol.PackHandshakeResponse41(h.clientHandshakeResponse)
	if err != nil {
		h.err = err
		return
	}

	tmp := packedHandshakeRespPacket[:(4+4+1+23)+4]
	sslPacket := make([]byte, len(tmp))
	copy(sslPacket, tmp)

	// This sets the payload_length bytes in the header portion of the packet
	// to reflect the new length of the truncated packet.
	pktLen := len(sslPacket) - 4
	sslPacket[0] = byte(pktLen)
	sslPacket[1] = byte(pktLen >> 8)
	sslPacket[2] = byte(pktLen >> 16)

	// Send TLS / SSL request packet
	h.writeBackendPacket(sslPacket)
	if h.err != nil {
		return
	}

	// Switch to TLS
	sslConn, err := ssl.HandleSSLUpgrade(h.backendConn.RawConnection(), *(h.dbSSLMode()))
	if err != nil {
		h.err = err
		return
	}
	h.backendConn.SetConnection(sslConn)
}

func (h *AuthenticationHandshake) writeClientHandshakeResponseToBackend() {
	if h.err != nil {
		return
	}

	// TODO: We should probably be carrying out a comprehensive unpacking, so that
	// we can be selective about the contents of the response
	packedHandshakeRespPacket, err := protocol.PackHandshakeResponse41(h.clientHandshakeResponse)
	if err != nil {
		h.err = err
		return
	}

	h.writeBackendPacket(packedHandshakeRespPacket)
}

func (h *AuthenticationHandshake) verifyAndProxyOkResponse() {
	if h.err != nil {
		return
	}

	// This proxying needs to take place to ensure the client gets the OK packet with
	// the correct sequence id, the connection keeps track of this information whereas
	// Secretless duplex streaming does not.
	rawPkt := h.readBackendPacket()
	h.writeClientPacket(rawPkt)
}

func (h *AuthenticationHandshake) handleBackendAuthResponse() {
	if h.err != nil {
		return
	}

	rawPkt := h.readBackendPacket()
	if h.err != nil {
		return
	}

	switch protocol.GetPacketType(rawPkt) {
	case protocol.ResponseAuthMoreData:
		defer h.verifyAndProxyOkResponse()

		moreDataResp, err := protocol.UnpackAuthMoreDataResponse(rawPkt)
		if err != nil {
			h.err = err
			return
		}

		switch moreDataResp.StatusTag {
		case protocol.CachingSha2PasswordFastAuthSuccess:
			// The user was cached and a fast login was performed successfully.
			// Do nothing. An OK packet will be sent by the server immediately
			// following this packet.
			return
		case protocol.CachingSha2PasswordPerformFullAuthentication:
			// The server is requesting a full authentication handshake.
			// https://dev.mysql.com/doc/dev/mysql-server/latest/page_caching_sha2_authentication_exchanges.html
			// https://github.com/go-sql-driver/mysql/blob/master/auth.go#L353

			// When using caching_sha2_password and TLS is enabled, no need
			// to fetch public key and sign password with it, since
			// the password is already encrypted in the TLS session.
			if h.clientRequestedSSL() {
				data, err := protocol.PackAuthSwitchResponse(
					h.backendConn.sequenceID,
					append([]byte(h.connectionDetails.Password), 0),
				)
				if err != nil {
					h.err = err
					return
				}

				h.writeBackendPacket(data)
				if h.err != nil {
					return
				}
				return
			}

			// Request public key from server
			data := protocol.PackAuthRequestPubKeyResponse(h.backendConn.sequenceID)

			h.writeBackendPacket(data)
			if h.err != nil {
				return
			}

			// Read public key from server
			pubKeyPkt := h.readBackendPacket()
			if h.err != nil {
				return
			}

			// Unpack public key from packet
			pubKey, err := protocol.UnpackAuthRequestPubKeyResponse(pubKeyPkt)
			if err != nil {
				h.err = err
				return
			}

			// Encrypt password with public key
			enc, err := protocol.EncryptPassword(h.connectionDetails.Password, h.serverHandshake.Salt, pubKey)
			if err != nil {
				h.err = err
				return
			}

			// Send encrypted password to server
			encPkt := protocol.PackAuthEncryptedPasswordResponse(h.backendConn.sequenceID, enc)

			h.writeBackendPacket(encPkt)
			return
		}

		return

	case protocol.ResponseAuthSwitchRequest:
		defer h.verifyAndProxyOkResponse()

		authSwitchRequest, err := protocol.UnpackAuthSwitchRequest(rawPkt)
		if err != nil {
			h.err = err
			return
		}

		salt := authSwitchRequest.PluginData
		// This is because the salt seems to actually be 21 bytes, ending in a null byte.
		// However the documentation suggests auth switch requests should be an EOF string
		// See https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase_packets_protocol_auth_switch_request.html
		if authSwitchRequest.PluginName == "mysql_native_password" {
			salt = salt[:20]
		}
		authResponse, err := protocol.CreateAuthResponse(authSwitchRequest.PluginName, []byte(h.connectionDetails.Password), salt)
		if err != nil {
			return
		}

		authSwitchResponseData, err := protocol.PackAuthSwitchResponse(
			authSwitchRequest.SequenceNumber,
			authResponse,
		)
		if err != nil {
			h.err = err
			return
		}
		h.writeBackendPacket(authSwitchResponseData)

		return

	default:
		// Let the client deal with it
		h.writeClientPacket(rawPkt)

		return
	}

}

func (h *AuthenticationHandshake) dbSSLMode() *ssl.DbSSLMode {
	if h.err != nil {
		return nil
	}
	// already memoized, just return it
	if h.sslMode != nil {
		return h.sslMode
	}

	var ret ssl.DbSSLMode
	ret, h.err = ssl.NewDbSSLMode(
		h.connectionDetails.SSLOptions, false,
	)
	h.sslMode = &ret

	return h.sslMode
}

func (h *AuthenticationHandshake) clientRequestedSSL() bool {
	if h.err != nil {
		return false
	}

	return h.dbSSLMode().UseTLS
}

func (h *AuthenticationHandshake) serverSupportsSSL() bool {
	return (h.serverHandshake.ServerCapabilities & protocol.ClientSSL) != 0
}

// NOTE: These lower level packet reading helper methods don't need the
//       h.err guards, becuase they'll always be called _by_ a higher
//       level method that has one.

func (h *AuthenticationHandshake) readClientPacket() Packet {
	return h.readPacket(h.clientConn)
}

func (h *AuthenticationHandshake) writeClientPacket(pkt Packet) {
	h.err = h.clientConn.write(pkt)
}

func (h *AuthenticationHandshake) readBackendPacket() Packet {
	return h.readPacket(h.backendConn)
}

func (h *AuthenticationHandshake) writeBackendPacket(pkt Packet) {
	h.err = h.backendConn.write(pkt)
}

// Just a helper method to DRY up the client/backend reads above
func (h *AuthenticationHandshake) readPacket(conn *Connection) Packet {
	pkt, err := conn.read()

	if err != nil {
		h.err = err
		return nil
	}

	return pkt
}
