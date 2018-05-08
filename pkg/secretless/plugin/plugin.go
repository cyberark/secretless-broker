package plugin

import (
	"net"

	"github.com/conjurinc/secretless/pkg/secretless"
)

// Plugin is an interface to be implemented by secretless plugins compiled as
// shared object files.
type Plugin interface {
	// Initialize is called before proxy initialization
	Initialize() error

	// CreateListener is called for every listener created by Proxy
	CreateListener(secretless.Listener)

	// NewConnection is called for each new client connection before being
	// passed to a handler
	NewConnection(secretless.Listener, net.Conn)

	// CloseConnect is called when a client connection is closed
	CloseConnection(net.Conn)

	// CreateHandler is called after listener creates a new handler
	CreateHandler(secretless.Listener, secretless.Handler)

	// DestroyHandler is called before a handler is removed
	DestroyHandler(secretless.Handler)

	// ResolveVariable is called when a provider resolves a variable
	ResolveVariable(p secretless.Provider, id string, value []byte)

	// ClientData is called for each inbound packet from clients
	ClientData([]byte)

	// ServerData is called for each inbound packet from the backend
	ServerData([]byte)
}
