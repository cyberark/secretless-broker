package example

import (
	"log"
	"net"
	"strconv"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// Manager is an empty struct
type Manager struct {
}

// Initialize is called before proxy initialization
func (manager *Manager) Initialize(c config.Config, configChangedFunc func(config.Config) error) error {
	log.Println("Initialized manager event...")
	return nil
}

// CreateListener is called for every listener created by Proxy
func (manager *Manager) CreateListener(l plugin_v1.Listener) {
	log.Println("Listener created manager event: ", l)
}

// NewConnection is called for each new client connection before being
// passed to a handler
func (manager *Manager) NewConnection(l plugin_v1.Listener, c net.Conn) {
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
func (manager *Manager) CloseConnection(c net.Conn) {
	log.Println("Close connection manager event...")
}

// CreateHandler is called after listener creates a new handler
func (manager *Manager) CreateHandler(h plugin_v1.Handler, c net.Conn) {
	log.Println("Create handler manager event...")
}

// DestroyHandler is called before a handler is removed
func (manager *Manager) DestroyHandler(h plugin_v1.Handler) {
	log.Println("Destroy handler manager event...")
}

// ResolveVariable is called when a provider resolves a variable
func (manager *Manager) ResolveVariable(provider plugin_v1.Provider, id string, value []byte) {
	log.Printf("Example-plugin ConnectionManager: Resolve variable manager event: %s = %s", id, string(value))
}

// ClientData is called for each inbound packet from clients
func (manager *Manager) ClientData(c net.Conn, buf []byte) {
	log.Println("Client data manager event...")
}

// ServerData is called for each inbound packet from the backend
func (manager *Manager) ServerData(c net.Conn, buf []byte) {
	log.Println("Server data manager event...")
}

// Shutdown is called before secretless exits
func (manager *Manager) Shutdown() {
	log.Println("Shutdown manager event...")
}

// ManagerFactory returns an empty Manager
func ManagerFactory() plugin_v1.ConnectionManager {
	return &Manager{}
}
