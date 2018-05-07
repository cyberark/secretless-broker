package mysql

import (
	"fmt"
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
	if values, err = variable.Resolve(h.Config.Credentials); err != nil {
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

	if values["schema"] != nil {
		result.Schema = string(values["schema"])
	}

	delete(values, "host")
	delete(values, "port")
	delete(values, "username")
	delete(values, "password")
	delete(values, "schema")

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

	// Proxy initial packet from server
	// TODO can we skip this step and still compute client packet?
	// how can we check the client accepts the protocol if we do?
	packet, err := protocol.ProxyPacket(backend, h.Client)
	if err != nil {
		return
	}

	// temp intercept of server packet
	// read server handshake
	//if _, err = protocol.ReadPacket(backend); err != nil {
	//	return
	//}
	//packet := []byte{74, 0, 0, 0, 10, 53, 46, 55, 46, 50, 49, 0, 195, 0, 2, 0, 115, 25, 43, 86, 114, 6, 120, 13, 0, 255, 255, 8, 2, 0, 255, 193, 21, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 52, 58, 27, 90, 29, 55, 50, 55, 4, 51, 3, 73, 0, 109, 121, 115, 113, 108, 95, 110, 97, 116, 105, 118, 101, 95, 112, 97, 115, 115, 119, 111, 114, 100, 0}
	//if _, err = protocol.WritePacket(packet, h.Client); err != nil {
	//	return
	//}

	fmt.Printf("initial server packet: %v\n", packet)

	// Decode server packet
	serverHandshake, err := protocol.DecodeHandshakeV10(packet)
	if err != nil {
		return
	}

	fmt.Printf("parsed server packet: %v\n", serverHandshake)

	// Intercept response from client
	interceptedClientPacket, err := protocol.ReadPacket(h.Client)
	if err != nil {
		return
	}

	fmt.Printf("initial client packet: %v\n", interceptedClientPacket)

	// Parse intercepted client response
	interceptedClientHandshake, err := protocol.DecodeHandshakeResponse41(interceptedClientPacket)
	if err != nil {
		return
	}

	fmt.Printf("parsed client packet: %v\n", interceptedClientHandshake)

	// Write client response with injected configuration
	clientPacket, err := protocol.GetHandshakeResponse41Packet(interceptedClientHandshake, serverHandshake, h.BackendConfig.Username, h.BackendConfig.Password)
	if err != nil {
		return
	}

	fmt.Printf("updated client packet: %v\n", clientPacket)

	// Send client response packet to server
	if _, err = protocol.WritePacket(clientPacket, backend); err != nil {
		return
	}

	// Proxy server response
	packet, err = protocol.ReadPacket(backend)
	if err != nil {
		return
	}

	OKResponse, err := protocol.DecodeOkResponse(packet)
	if err != nil {
		return
	}

	fmt.Printf("server OK response: %v\n", OKResponse)

	// Proxy OK packet to client
	if _, err = protocol.WritePacket(packet, h.Client); err != nil {
		return
	}

	//	if packet[4] == protocol.ResponseErr {
	//		err = protocol.ParseError(packet)
	//		return
	//	}

	if h.Config.Debug {
		log.Printf("Successfully connected to '%s:%d'", h.BackendConfig.Host, h.BackendConfig.Port)
	}

	//if _, err = h.Client.Write(protocol.CreateAuthenticationOKMessage()); err != nil {
	//	return
	//}

	h.Backend = backend

	return
}
