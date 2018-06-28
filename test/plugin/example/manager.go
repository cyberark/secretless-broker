package example

import (
	"log"
	"net"
	"strconv"

	"github.com/conjurinc/secretless/pkg/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
)

type ExampleManager struct {
}

// Initialize is called before proxy initialization
func (manager *ExampleManager) Initialize(c config.Config) error {
	log.Println("Initialized manager event...")
	return nil
}

// CreateListener is called for every listener created by Proxy
func (manager *ExampleManager) CreateListener(l plugin_v1.Listener) {
	log.Println("Listener created manager event: ", l)
}

// NewConnection is called for each new client connection before being
// passed to a handler
func (manager *ExampleManager) NewConnection(l plugin_v1.Listener, c net.Conn) {
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
func (manager *ExampleManager) CloseConnection(c net.Conn) {
	log.Println("Close connection manager event...")
}

// CreateHandler is called after listener creates a new handler
func (manager *ExampleManager) CreateHandler(h plugin_v1.Handler, c net.Conn) {
	log.Println("Create handler manager event...")
}

// DestroyHandler is called before a handler is removed
func (manager *ExampleManager) DestroyHandler(h plugin_v1.Handler) {
	log.Println("Destroy handler manager event...")
}

// ResolveVariable is called when a provider resolves a variable
func (manager *ExampleManager) ResolveVariable(provider secretless.Provider, id string, value []byte) {
	log.Println("Resolve variable manager event...")
}

// ClientData is called for each inbound packet from clients
func (manager *ExampleManager) ClientData(c net.Conn, buf []byte) {
	log.Println("Client data manager event...")
}

// ServerData is called for each inbound packet from the backend
func (manager *ExampleManager) ServerData(c net.Conn, buf []byte) {
	log.Println("Server data manager event...")
}

// Shutdown is called before secretless exits
func (manager *ExampleManager) Shutdown() {
	log.Println("Shutdown manager event...")
}

func ManagerFactory() plugin_v1.ConnectionManager {
	return &ExampleManager{}
}
