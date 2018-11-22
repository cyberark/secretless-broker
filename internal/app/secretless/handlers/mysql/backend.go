package mysql

import (
	"crypto/tls"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/mysql/protocol"
)

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{Options: make(map[string]string)}

	var values map[string][]byte
	if values, err = h.Resolver.Resolve(h.GetConfig().Credentials); err != nil {
		return
	}

	if h.GetConfig().Debug {
		log.Printf("MySQL backend connection parameters: %s", values)
	}

	if host := values["host"]; host != nil {
		result.Host = string(values["host"])
	}

	if values["port"] != nil {
		port64, _ := strconv.ParseUint(string(values["port"]), 10, 64)
		result.Port = uint(port64)
	}

	if values["username"] != nil {
		result.Username = string(values["username"])
	}

	if values["password"] != nil {
		result.Password = string(values["password"])
	}

	delete(values, "host")
	delete(values, "port")
	delete(values, "username")
	delete(values, "password")

	result.Options = make(map[string]string)
	for k, v := range values {
		result.Options[k] = string(v)
	}

	h.BackendConfig = &result

	return
}

// ConnectToBackend establishes the connection to the backend database and sets the Backend field.
func (h *Handler) ConnectToBackend() (err error) {
	var backend net.Conn

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
	packet, err := protocol.ProxyPacket(backend, h.GetClientConnection())
	if err != nil {
		return
	}

	// Unpack server packet
	serverHandshake, err := protocol.UnpackHandshakeV10(packet)
	if err != nil {
		return
	}

	// TODO: if SSL requested but server doesn't support FAIL
	// serverHandshake.ServerCapabilities&protocol.ClientSSL > 0

	// Intercept response from client
	handshakeResponsePacket, err := protocol.ReadPacket(h.GetClientConnection())
	if err != nil {
		return
	}
	// Track clientSequenceID
	clientSequenceID := handshakeResponsePacket[3]

	// Parse intercepted client response
	handshakeResponse, err := protocol.UnpackHandshakeResponse41(handshakeResponsePacket)
	if err != nil {
		return
	}

	// TODO: handle other scenarios beyond skip-verify by providing configuration tls
	requestedSSL := strings.ToLower(h.BackendConfig.Options["sslmode"]) == "required"

	// Ensure CapabilityFlag is set when using TLS
	if requestedSSL {
		handshakeResponse.CapabilityFlags |= protocol.ClientSSL
	}

	// Inject credentials into client response
	if err = protocol.InjectCredentials(handshakeResponse, serverHandshake.Salt, h.BackendConfig.Username, h.BackendConfig.Password); err != nil {
		return
	}

	// Pack client response with injected configuration
	packedHandshakeRespPacket, err := protocol.PackHandshakeResponse41(handshakeResponse)

	if err != nil {
		return
	}

	if requestedSSL {
		// Send TLS / SSL request packet

		// Copy a truncated HandshakeResponse to create SSLRequest
		// https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::SSLRequest
		tmp := packedHandshakeRespPacket[:(4+4+1+23)+4]
		sslPacket := make([]byte, len(tmp))
		copy(sslPacket, tmp)

		// Update packet length for truncated packet
		pktLen := len(sslPacket) - 4
		sslPacket[0] = byte(pktLen)
		sslPacket[1] = byte(pktLen >> 8)
		sslPacket[2] = byte(pktLen >> 16)

		// Send TLS / SSL request packet
		if _, err = protocol.WritePacket(sslPacket, backend); err != nil {
			return
		}
		// Increment sequenceID in anticipation of subsequent write to backend
		packedHandshakeRespPacket[3]++;

		// Switch to TLS
		tlsClient := tls.Client(backend, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err := tlsClient.Handshake(); err != nil {
			return err
		}
		backend = tlsClient
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
