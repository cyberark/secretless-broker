package eventnotifier

import (
	"net"

	v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

type defaultNotifier struct {
	connectionManagers []v1.ConnectionManager
}

// New returns an EventNotifier that delegates all event notifications to the
// ConnectionManagers it's been started with
func New(cxnManagers []v1.ConnectionManager) v1.EventNotifier {
	return &defaultNotifier{
		connectionManagers: cxnManagers,
	}
}

// NewConnection loops through the connection managers and adds a connection to the listener l
// ResolveCredential loops through the connection managers to resolve the secret specified
func (dn *defaultNotifier) ResolveCredential(provider v1.Provider, id string, value []byte) {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.ResolveCredential(provider, id, value)
	}
}

// ClientData loops through the connection managers to proxy data from the client
func (dn *defaultNotifier) ClientData(c net.Conn, buf []byte) {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.ClientData(c, buf)
	}
}

// ServerData loops through the connection managers to proxy data from the server
func (dn *defaultNotifier) ServerData(c net.Conn, buf []byte) {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.ServerData(c, buf)
	}
}

// TODO: The two methods below are currently NOT on the EventNotifier
//   interface, but are on the ConnectionManager interface -- should they
//   be on EventNotifier?

// Shutdown calls Shutdown on the Proxy and all the connection managers, concurrently
func (dn *defaultNotifier) Shutdown() {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.Shutdown()
	}
}

// CloseConnection loops through the connection managers and closes the connection c
func (dn *defaultNotifier) CloseConnection(c net.Conn) {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.CloseConnection(c)
	}
}
