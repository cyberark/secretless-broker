package v1

import (
	"net"

	"github.com/conjurinc/secretless-broker/pkg/secretless/config"
)

// ConnectionManager is an interface to be implemented by plugins that want to
// manage connections for handlers and listeners.
type ConnectionManager interface {
	// Initialize is called before proxy initialization
	Initialize(config.Config, func(config.Config) error) error

	// CreateListener is called for every listener created by Proxy
	CreateListener(Listener)

	// NewConnection is called for each new client connection before being
	// passed to a handler
	NewConnection(Listener, net.Conn)

	// CloseConnect is called when a client connection is closed
	CloseConnection(net.Conn)

	// CreateHandler is called after listener creates a new handler
	CreateHandler(Handler, net.Conn)

	// DestroyHandler is called before a handler is removed
	DestroyHandler(Handler)

	// ResolveVariable is called when a provider resolves a variable
	ResolveVariable(provider Provider, id string, value []byte)

	// ClientData is called for each inbound packet from clients
	ClientData(net.Conn, []byte)

	// ServerData is called for each inbound packet from the backend
	ServerData(net.Conn, []byte)

	// Shutdown is called when secretless caught a signal to exit
	Shutdown()
}
