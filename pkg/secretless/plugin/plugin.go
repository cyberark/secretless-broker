package plugin

import (
	"net"

	"github.com/conjurinc/secretless/pkg/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/conjurinc/secretless/pkg/secretless/handler"
	"github.com/conjurinc/secretless/pkg/secretless/listener"
)

// Plugin is an interface to be implemented by secretless plugins compiled as
// shared object files.
type Plugin interface {
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

// ----------------- V1 interfaces -----------------
// Listener v1 interface
type Listener_v1 = listener.Listener_v1

// Handler v1 interface
type Handler_v1 = handler.Handler_v1

// Manager v1 interface
// TODO: Move this to its own folder
// TODO: Really define this interface
type Manager_v1 = Plugin
