package example

import (
	"log"
	"net"
	"strconv"

	config_v1 "github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// connectionManager is an empty struct
type connectionManager struct {
}

// Initialize is called before proxy initialization
func (manager *connectionManager) Initialize(c config_v1.Config, configChangedFunc func(config_v1.Config) error) error {
	log.Println("Initialized manager event...")
	return nil
}

// CreateListener is called for every listener created by Proxy
func (manager *connectionManager) CreateListener(l plugin_v1.Listener) {
	log.Println("Listener created manager event: ", l)
}

// NewConnection is called for each new client connection before being
// passed to a handler
func (manager *connectionManager) NewConnection(l plugin_v1.Listener, c net.Conn) {
	_, port, err := net.SplitHostPort(c.LocalAddr().String())
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("New connection manager event on port %s...", port)

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalln(err)
	}

	if portNumber%2 == 0 {
		log.Printf("Connection on %v blocked since it's an even port number", portNumber)
		c.Close()
		return
	}

	log.Printf("Connection on %v allowed since it's on an odd port number", portNumber)
}

// CloseConnection is called when a client connection is closed
func (manager *connectionManager) CloseConnection(c net.Conn) {
	log.Println("Close connection manager event...")
}

// CreateHandler is called after listener creates a new handler
func (manager *connectionManager) CreateHandler(h plugin_v1.Handler, c net.Conn) {
	log.Println("Create handler manager event...")
}

// DestroyHandler is called before a handler is removed
func (manager *connectionManager) DestroyHandler(h plugin_v1.Handler) {
	log.Println("Destroy handler manager event...")
}

// ResolveSecret is called when a provider resolves a variable
func (manager *connectionManager) ResolveSecret(provider plugin_v1.Provider, id string, value []byte) {
	log.Printf("Example-plugin ConnectionManager: Resolve secret manager event: %s = %s", id, string(value))
}

// ClientData is called for each inbound packet from clients
func (manager *connectionManager) ClientData(c net.Conn, buf []byte) {
	log.Println("Client data manager event...")
}

// ServerData is called for each inbound packet from the backend
func (manager *connectionManager) ServerData(c net.Conn, buf []byte) {
	log.Println("Server data manager event...")
}

// Shutdown is called before secretless exits
func (manager *connectionManager) Shutdown() {
	log.Println("Shutdown manager event...")
}

// ConnManagerFactory returns an empty ConnectionManager
func ConnManagerFactory() plugin_v1.ConnectionManager {
	return &connectionManager{}
}
