package mysql

import (
	"errors"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/ssl"
	"log"
	"net"
	"strconv"

	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/mysql/protocol"
)

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
		log.Printf("MySQL backend connection parameters: %s", connectionDetails)
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
	var backend net.Conn

	// resolve TLS Configuration from BackendConfig Options
	tlsConf, err := ssl.ResolveTLSConfig(h.BackendConfig.Options, false)
	requestedSSL := tlsConf.UseTLS
	if err != nil {
		return
	}

	address := h.BackendConfig.Host + ":" + strconv.FormatUint(uint64(h.BackendConfig.Port), 10)

	if backend, err = net.Dial("tcp", address); err != nil {
		return
	}

	if h.GetConfig().Debug {
		log.Print("Processing handshake")
	}

	// Proxy initial packet from server to client
	// TODO can we skip this step and still compute client packet?
	// how can we check the client accepts the protocol if we do?
	packet, err := protocol.ReadPacket(backend)
	if err != nil {
		return err
	}

	// Remove Client SSL Capability from Server Handshake Packet
	packetWithNoSSL, err := protocol.RemoveSSLFromHandshakeV10(packet)
	if err != nil {
		return err
	}

	_, err = protocol.WritePacket(packetWithNoSSL, h.GetClientConnection())
	if err != nil {
		return err
	}

	// Unpack server packet
	serverHandshake, err := protocol.UnpackHandshakeV10(packet)
	if err != nil {
		return
	}

	// requested SSL but server doesn't support it
	if requestedSSL && serverHandshake.ServerCapabilities&protocol.ClientSSL == 0 {
		return &protocol.Error{
			Code:     protocol.CRSSLConnectionError,
			SQLSTATE: protocol.ErrorCodeInternalError,
			Message:  ErrNoTLS.Error(),
			SequenceID: 2,
		}
	}

	// Intercept response from client
	handshakeResponsePacket, err := protocol.ReadPacket(h.GetClientConnection())
	if err != nil {
		return
	}
	// Track clientSequenceID
	clientSequenceID := handshakeResponsePacket[3]

	// Parse intercepted client response
	// TODO: client requesting SSL results ERROR 2026 (HY000): SSL connection error: protocol version mismatch
	handshakeResponse, err := protocol.UnpackHandshakeResponse41(handshakeResponsePacket)
	if err != nil {
		return
	}

	// Ensure CapabilityFlag is set when using TLS
	if requestedSSL {
		handshakeResponse.CapabilityFlags |= protocol.ClientSSL
	}

	// Inject credentials into client response
	if err = protocol.InjectCredentials(handshakeResponse, serverHandshake.Salt, h.BackendConfig.Username, h.BackendConfig.Password); err != nil {
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
	packedHandshakeRespPacket, err := protocol.PackHandshakeResponse41(handshakeResponse)

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

		// Send TLS / SSL request packet
		if _, err = protocol.WritePacket(sslPacket, backend); err != nil {
			return
		}
		// Increment sequenceID in anticipation of subsequent write to backend
		packedHandshakeRespPacket[3]++

		// Switch to TLS
		backend, err = ssl.HandleSSLUpgrade(backend, tlsConf)
		if err != nil {
			return
		}
	}

	// Send configured client response packet to server
	if _, err = protocol.WritePacket(packedHandshakeRespPacket, backend); err != nil {
		return
	}

	// Proxy server response
	packet, err = protocol.ReadPacket(backend)
	if err != nil {
		return
	}

	_, err = protocol.UnpackOkResponse(packet)
	if err != nil {
		return
	}

	// Proxy OK packet to client
	// TODO: refactor to wrap conn in struct that can
	// keep track of clientSequenceID
	// be custodian of client connection logic
	// https://github.com/siddontang/mixer/blob/master/proxy/conn.go
	packet[3] = clientSequenceID + 1 // ensure usage of incremented clientSequenceID
	if _, err = protocol.WritePacket(packet, h.GetClientConnection()); err != nil {
		return
	}

	if h.GetConfig().Debug {
		log.Printf("Successfully connected to '%s:%d'", h.BackendConfig.Host, h.BackendConfig.Port)
	}

	h.BackendConnection = backend

	return
}
