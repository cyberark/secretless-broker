package v1

import (
	"net"
)

// EventNotifier is the interface which is used to pass event up from handlers/
// listeners/managers back up to the main plugin manager
type EventNotifier interface {
	// ClientData is called for each inbound packet from clients
	ClientData(net.Conn, []byte)

	// ResolveCredential is called when a provider resolves a variable
	// TODO: unclear why we're reimplementing the StoredSecret functionality here...
	ResolveCredential(provider Provider, id string, value []byte)

	// ServerData is called for each inbound packet from the backend
	ServerData(net.Conn, []byte)
}
