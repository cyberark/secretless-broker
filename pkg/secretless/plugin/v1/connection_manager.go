package v1

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	"net"
)

// ConnectionManager is an interface to be implemented by plugins that want to
// manage connections for handlers and listeners.
type ConnectionManager interface {
	// Initialize is called before proxy initialization
	Initialize(v1.Config, func(v1.Config) error) error

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

	// ResolveSecret is called when a provider resolves a variable
	ResolveSecret(provider Provider, id string, value []byte)

	// ClientData is called for each inbound packet from clients
	ClientData(net.Conn, []byte)

	// ServerData is called for each inbound packet from the backend
	ServerData(net.Conn, []byte)

	// Shutdown is called when secretless caught a signal to exit
	Shutdown()
}
