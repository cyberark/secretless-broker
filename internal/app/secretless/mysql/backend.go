package mysql

import (
	"log"
	"net"
	"strconv"

	"github.com/conjurinc/secretless/internal/app/secretless/mysql/protocol"
	"github.com/conjurinc/secretless/internal/app/secretless/variable"
)

// ConfigureBackend resolves the backend connection settings and credentials and sets the
// BackendConfig field.
func (h *Handler) ConfigureBackend() (err error) {
	result := BackendConfig{Options: make(map[string]string)}

	var values map[string][]byte
	if values, err = variable.Resolve(h.Config.Credentials, h.EventNotifier); err != nil {
		return
	}

	if h.Config.Debug {
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

	if h.Config.Debug {
		log.Print("Processing handshake")
	}

	// Proxy initial packet from server to client
	// TODO can we skip this step and still compute client packet?
	// how can we check the client accepts the protocol if we do?
	packet, err := protocol.ProxyPacket(backend, h.Client)
	if err != nil {
		return
	}

	// Unpack server packet
	serverHandshake, err := protocol.UnpackHandshakeV10(packet)
	if err != nil {
		return
	}

	// Intercept response from client
	handshakeResponsePacket, err := protocol.ReadPacket(h.Client)
	if err != nil {
		return
	}

	// Parse intercepted client response
	handshakeResponse, err := protocol.UnpackHandshakeResponse41(handshakeResponsePacket)
	if err != nil {
		return
	}

	// Inject credentials into client response
	if err = protocol.InjectCredentials(handshakeResponse, serverHandshake.Salt, h.BackendConfig.Username, h.BackendConfig.Password); err != nil {
		return
	}

	// Pack client response with injected configuration
	clientPacket, err := protocol.PackHandshakeResponse41(handshakeResponse)
	if err != nil {
		return
	}

	// Send configured client response packet to server
	if _, err = protocol.WritePacket(clientPacket, backend); err != nil {
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
	if _, err = protocol.WritePacket(packet, h.Client); err != nil {
		return
	}

	if h.Config.Debug {
		log.Printf("Successfully connected to '%s:%d'", h.BackendConfig.Host, h.BackendConfig.Port)
	}

	h.Backend = backend

	return
}
