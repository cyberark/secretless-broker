package mysql

import (
	"errors"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/ssl"
	"log"
	"net"
	"reflect"
	"strconv"

	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/mysql/protocol"
)

// TODO: move to another file? perhaps into packet?
//
type mysqlPacket []byte

func (pkt *mysqlPacket) SequenceId() byte {
	return (*pkt)[3]
}

func (pkt *mysqlPacket) SetSequenceId(id byte) {
	(*pkt)[3] = id
}

// Various errors the handler might return
var (
	ErrNoTLS = errors.New("SSL connection error: SSL is required but the server doesn't support it")
)

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{Options: make(map[string]string)}

	var connectionDetails map[string][]byte
	if connectionDetails, err = h.Resolver.Resolve(h.GetConfig().Credentials); err != nil {
		return
	}

	if h.GetConfig().Debug {
		keys := reflect.ValueOf(connectionDetails).MapKeys()
		log.Printf("%s backend connection parameters: %s", h.GetConfig().Name, keys)
	}

	if host := connectionDetails["host"]; host != nil {
		result.Host = string(connectionDetails["host"])
	}

	if connectionDetails["port"] != nil {
		port64, _ := strconv.ParseUint(string(connectionDetails["port"]), 10, 64)
		result.Port = uint(port64)
	}

	if connectionDetails["username"] != nil {
		result.Username = string(connectionDetails["username"])
	}

	if connectionDetails["password"] != nil {
		result.Password = string(connectionDetails["password"])
	}

	delete(connectionDetails, "host")
	delete(connectionDetails, "port")
	delete(connectionDetails, "username")
	delete(connectionDetails, "password")

	result.Options = make(map[string]string)
	for k, v := range connectionDetails {
		result.Options[k] = string(v)
	}

	h.BackendConfig = &result

	return
}


// ConnectToBackend establishes the connection to the backend database and sets the Backend field.
func (h *Handler) ConnectToBackend() (err error) {

	h.ResetSequenceIds()

	// resolve TLS Configuration from BackendConfig Options
	tlsConf, err := ssl.NewSecretlessTLSConfig(h.BackendConfig.Options, false)
	if err != nil {
		return
	}
	requestedSSL := tlsConf.UseTLS

	err = h.OpenBackendConnection()
	if err != nil {
		return
	}

	h.PrintDebug("Processing handshake")

	// STEP: ServerGreeting. Backend => Secretless
	//////////////////////////////////////////////

	// Proxy initial packet from server to client
	// TODO can we skip this step and still compute client packet?
	// how can we check the client accepts the protocol if we do?
	serverHandshakePacket, err := h.BackendRead()
	if err != nil {
		return err
	}

	// STEP: ServerGreeting. Secretless => Client
	//////////////////////////////////////////////

	// Unpack server packet
	serverHandshake, err := protocol.NewHandshakeV10(serverHandshakePacket)
	serverDoesntSupportSSL := !serverHandshake.SupportsSSL()
	if err != nil {
		return
	}

	serverHandshake.RemoveClientSSL()
	serverHandshakePacketWithNoSSL, err := serverHandshake.Pack()
	if err != nil {
		return err
	}

	// Write Handshake to Client
	_, err = h.ClientWrite(serverHandshakePacketWithNoSSL)
	if err != nil {
		return err
	}

	// STEP: LoginRequest. Client => Secretless
	//////////////////////////////////////////////

	// Intercept response from client
	clientHandshakeResponsePacket, err := h.ClientRead()
	if err != nil {
		return err
	}

	if requestedSSL && serverDoesntSupportSSL {
		return &protocol.Error{
			Code:       protocol.CRSSLConnectionError,
			SQLSTATE:   protocol.ErrorCodeInternalError,
			Message:    ErrNoTLS.Error(),
		}
	}

	// Parse intercepted client response
	// TODO: client requesting SSL results ERROR 2026 (HY000): SSL connection error: protocol version mismatch

	clientHandshakeResponse, err := protocol.NewHandshakeResponse41(clientHandshakeResponsePacket)
	if err != nil {
		return err
	}

	// TODO: add tests cases for authentication plugins support
	// Disable CapabilityFlag for authentication plugins support
	clientHandshakeResponse.DisableClientPluginAuth()

	// TODO: add tests cases for client secure connection
	// Enable CapabilityFlag for client secure connection
	clientHandshakeResponse.EnableClientSecureConnection()

	// Ensure CapabilityFlag is set when using TLS
	if requestedSSL {
		clientHandshakeResponse.EnableSSL()
	}

	// Inject credentials into client response
	if err = clientHandshakeResponse.InjectCredentials(serverHandshake.Salt(), h.BackendConfig.Username, h.BackendConfig.Password); err != nil {
		return
	}

	// This format of this packet is described here:
	//
	//   https://dev.mysql.com/doc/internals/en/mysql-packet.html
	//
	//  +-------------+----------------+---------------------------------------------+
	//  |    Type     |      Name      |                 Description                 |
	//	+-------------+----------------+---------------------------------------------+
	//  | int<3>      | payload_length | Length of the payload. The number of bytes  |
	//  |             |                | in the packet beyond the initial 4 bytes    |
	//  |             |                | that make up the packet header.             |
	//  | int<1>      | sequence_id    | Sequence ID                                 |
	//  | string<var> | payload        | [len=payload_length] payload of the packet  |
	//  +-------------+----------------+---------------------------------------------+
	packedHandshakeRespPacket, err := clientHandshakeResponse.Pack()

	if err != nil {
		return
	}

	// Send TLS / SSL request packet
	//
	if requestedSSL {
		// The SSLRequest packet is created by copying the HandshakeResponse,
		// but truncating the username and everything after the username in
		// the payload, as described here:
		//
		// https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::SSLRequest
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
		//
		tmp := packedHandshakeRespPacket[:(4+4+1+23)+4]
		sslPacket := make([]byte, len(tmp))
		copy(sslPacket, tmp)

		// This sets the payload_length bytes in the header portion of the packet
		// to reflect the new length of the truncated packet.
		pktLen := len(sslPacket) - 4
		sslPacket[0] = byte(pktLen)
		sslPacket[1] = byte(pktLen >> 8)
		sslPacket[2] = byte(pktLen >> 16)

// STEP: SSLRequestModifiedLoginRequest. Secretless => Backend
		// Send TLS / SSL request packet
		if _, err = h.BackendWrite(sslPacket); err != nil {
			return
		}

		// Switch to TLS
// STEP: TLSUpgrade. Secretless => Backend
		h.BackendConnection, err = ssl.HandleSSLUpgrade(h.BackendConnection, tlsConf)
		if err != nil {
			return
		}
	}

// STEP: LoginRequest. Secretless => Backend
	// Send configured client response packet to server
	if _, err = h.BackendWrite(packedHandshakeRespPacket); err != nil {
		return
	}

// TODO: AuthSwitchRequest, AuthSwitchResponse
// https://dev.mysql.com/doc/refman/8.0/en/authentication-plugins.html


// STEP: LoginResponse. Backend => Secretless
	// Proxy server response
	loginResponsePacket, err := h.BackendRead();
	if err != nil {
		return
	}

	if _, err = protocol.NewOkResponse(loginResponsePacket); err != nil {
		return
	}

// STEP: LoginResponse. Secretless => Client
	// Proxy OK packet to client
	// TODO: refactor to wrap conn in struct that can
	// keep track of clientSequenceID
	// be custodian of client connection logic
	// https://github.com/siddontang/mixer/blob/master/proxy/conn.go
	if _, err = h.ClientWrite(loginResponsePacket); err != nil {
		return
	}

	if h.GetConfig().Debug {
		log.Printf("Successfully connected to '%s:%d'", h.BackendConfig.Host, h.BackendConfig.Port)
	}

	return
}

// Read and Write
func (h *Handler) BackendWrite(pkt mysqlPacket) (int, error) {
	h.CopyBackendSequenceIdIntoPacket(&pkt)
	h.IncrementBackendSequenceId()
	return protocol.WritePacket(pkt, h.GetBackendConnection())
}

func (h *Handler) BackendRead() ([]byte, error) {
	pkt, err := protocol.ReadPacket(h.GetBackendConnection())
	if err != nil {
		return nil, err
	}
	h.SetBackendSequenceIdFromPacket(pkt)
	h.IncrementBackendSequenceId()
	return pkt, err
}

func (h *Handler) ClientWrite(pkt mysqlPacket) (int, error) {
	h.CopyClientSequenceIdIntoPacket(&pkt)
	h.IncrementClientSequenceId()
	return protocol.WritePacket(pkt, h.GetClientConnection())
}

func (h *Handler) ClientRead() ([]byte, error) {
	pkt, err := protocol.ReadPacket(h.GetClientConnection())
	if err != nil {
		return nil, err
	}
	//h.clientSequenceId = pkt[3] + 1
	// TODO verify: is below line same as above line?
	// we shouldn't need to re-read the sequence id from the packet,
	// because it's already stored
	h.IncrementClientSequenceId()
	return pkt, err
}

func (h *Handler) ResetSequenceIds() {
	h.clientSequenceId = 0 // starts at with ServerGreeting or ServerGreetingError
	h.backendSequenceId = 1 // start at 1 with LoginRequest or SSLRequestModifiedLoginRequest
}

func (h *Handler) IncrementClientSequenceId() {
	h.clientSequenceId++
}

func (h *Handler) IncrementBackendSequenceId() {
	h.backendSequenceId++
}

func (h *Handler) CopyClientSequenceIdIntoPacket(pkt *mysqlPacket) {
	pkt.SetSequenceId(h.clientSequenceId)
}

func (h *Handler) CopyBackendSequenceIdIntoPacket(pkt *mysqlPacket) {
	pkt.SetSequenceId(h.backendSequenceId)
}

func (h *Handler) SetBackendSequenceIdFromPacket(pkt mysqlPacket) {
	h.backendSequenceId = pkt.SequenceId()
}
func (h *Handler) OpenBackendConnection() (err error) {
	h.BackendConnection, err = net.Dial("tcp", h.BackendConfig.Address())
	return err
}
