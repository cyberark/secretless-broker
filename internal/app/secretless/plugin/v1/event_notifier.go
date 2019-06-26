package v1

import (
	"net"
)

// EventNotifier is the interface which is used to pass event up from handlers/
// listeners/managers back up to the main plugin manager
type EventNotifier interface {
	// NewConnection is called for each new client connection before being
	// passed to a handler
	NewConnection(Listener, net.Conn)

	// ClientData is called for each inbound packet from clients
	ClientData(net.Conn, []byte)

	// CreateHandler is called after listener creates a new handler
	CreateHandler(Handler, net.Conn)

	// CreateListener is called for every listener created by Proxy
	CreateListener(Listener)

	// ResolveSecret is called when a provider resolves a variable
	// TODO: unclear why we're reimplementing the StoredSecret functionality here...
	ResolveSecret(provider Provider, id string, value []byte)

	// ServerData is called for each inbound packet from the backend
	ServerData(net.Conn, []byte)
}
