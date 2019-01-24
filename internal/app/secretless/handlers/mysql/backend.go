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


// Read and Write
func (h *Handler) BackendWrite(pkt []byte) (int, error) {
	// use current backendSequenceId onWrite
	pkt[3] = h.backendSequenceId
	n, err := protocol.WritePacket(pkt, h.GetBackendConnection())

	// increment backendSequenceId in anticipation of subsequent write
	h.backendSequenceId++

	return n, err
}

func (h *Handler) BackendRead() ([]byte, error) {
	pkt, err := protocol.ReadPacket(h.GetBackendConnection())
	if err == nil {
		// increment backendSequenceId in anticipation of subsequent write
		h.backendSequenceId = pkt[3] + 1
	}
	return pkt, err
}

func (h *Handler) ClientWrite(pkt []byte) (int, error) {
	// use current clientSequenceId onWrite
	pkt[3] = h.clientSequenceId
	n, err := protocol.WritePacket(pkt, h.GetClientConnection())

	// increment clientSequenceId in anticipation of subsequent write
	h.clientSequenceId++

	return n, err
}

func (h *Handler) ClientRead() ([]byte, error) {
	pkt, err := protocol.ReadPacket(h.GetClientConnection())
	if err == nil {
		// increment clientSequenceId in anticipation of subsequent write
		h.clientSequenceId = pkt[3] + 1
	}
	return pkt, err
}

// ConnectToBackend establishes the connection to the backend database and sets the Backend field.
func (h *Handler) ConnectToBackend() (err error) {

	h.clientSequenceId = 0 // starts at with ServerGreeting or ServerGreetingError
	h.backendSequenceId = 1 // start at 1 with LoginRequest or SSLRequestModifiedLoginRequest

	// resolve TLS Configuration from BackendConfig Options
	tlsConf, err := ssl.NewSecretlessTLSConfig(h.BackendConfig.Options, false)
	requestedSSL := tlsConf.UseTLS
	if err != nil {
		return
	}

	address := h.BackendConfig.Host + ":" + strconv.FormatUint(uint64(h.BackendConfig.Port), 10)

	h.BackendConnection, err = net.Dial("tcp", address)
	if err != nil {
		return
	}

	if h.GetConfig().Debug {
		log.Print("Processing handshake")
	}

// STEP: ServerGreeting. Backend => Secretless
	// Proxy initial packet from server to client
	// TODO can we skip this step and still compute client packet?
	// how can we check the client accepts the protocol if we do?
	packet, err := h.BackendRead()
	if err != nil {
		return err
	}

	// Remove Client SSL Capability from Server Handshake Packet
	packetWithNoSSL, err := protocol.RemoveSSLFromHandshakeV10(packet)
	if err != nil {
		return err
	}

// STEP: ServerGreeting. Secretless => Client
	_, err = h.ClientWrite(packetWithNoSSL)
	if err != nil {
		return err
	}

	// Unpack server packet
	serverHandshake, err := protocol.UnpackHandshakeV10(packet)
	if err != nil {
		return
	}

// STEP: LoginRequest. Client => Secretless
	// Intercept response from client
	handshakeResponsePacket, err := h.ClientRead()

	// requested SSL but server doesn't support it
	if requestedSSL && serverHandshake.ServerCapabilities&protocol.ClientSSL == 0 {
		return &protocol.Error{
			Code:       protocol.CRSSLConnectionError,
			SQLSTATE:   protocol.ErrorCodeInternalError,
			Message:    ErrNoTLS.Error(),
		}
	}

	// Parse intercepted client response
	// TODO: client requesting SSL results ERROR 2026 (HY000): SSL connection error: protocol version mismatch
	handshakeResponse, err := protocol.UnpackHandshakeResponse41(handshakeResponsePacket)
	if err != nil {
		return &protocol.Error{
			Code:       protocol.CRSSLConnectionError,
			SQLSTATE:   protocol.ErrorCodeInternalError,
			Message:    ErrNoTLS.Error(),
		}
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
	if packet, err = h.BackendRead(); err != nil {
		return
	}

	if _, err = protocol.UnpackOkResponse(packet); err != nil {
		return
	}

// STEP: LoginResponse. Secretless => Client
	// Proxy OK packet to client
	// TODO: refactor to wrap conn in struct that can
	// keep track of clientSequenceID
	// be custodian of client connection logic
	// https://github.com/siddontang/mixer/blob/master/proxy/conn.go
	if _, err = h.ClientWrite(packet); err != nil {
		return
	}

	if h.GetConfig().Debug {
		log.Printf("Successfully connected to '%s:%d'", h.BackendConfig.Host, h.BackendConfig.Port)
	}

	return
}
