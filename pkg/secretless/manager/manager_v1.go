package manager

import (
	"net"

	"github.com/conjurinc/secretless/pkg/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/conjurinc/secretless/pkg/secretless/handler"
	"github.com/conjurinc/secretless/pkg/secretless/listener"
)

// Manager is an interface to be implemented by plugins that want to
// manage connections for handlers and listeners.
type Manager_v1 interface {
	// Initialize is called before proxy initialization
	Initialize(config.Config) error

	// CreateListener is called for every listener created by Proxy
	CreateListener(listener.Listener_v1)

	// NewConnection is called for each new client connection before being
	// passed to a handler
	NewConnection(listener.Listener_v1, net.Conn)

	// CloseConnect is called when a client connection is closed
	CloseConnection(net.Conn)

	// CreateHandler is called after listener creates a new handler
	CreateHandler(handler.Handler_v1, net.Conn)

	// DestroyHandler is called before a handler is removed
	DestroyHandler(handler.Handler_v1)

	// ResolveVariable is called when a provider resolves a variable
	ResolveVariable(p secretless.Provider, id string, value []byte)

	// ClientData is called for each inbound packet from clients
	ClientData(net.Conn, []byte)

	// ServerData is called for each inbound packet from the backend
	ServerData(net.Conn, []byte)

	// Shutdown is called when secretless caught a signal to exit
	Shutdown()
}
