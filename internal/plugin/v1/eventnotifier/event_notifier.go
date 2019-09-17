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

func (dn *defaultNotifier) CreateListener(l v1.Listener) {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.CreateListener(l)
	}
}

// NewConnection loops through the connection managers and adds a connection to the listener l
func (dn *defaultNotifier) NewConnection(l v1.Listener, c net.Conn) {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.NewConnection(l, c)
	}
}

// CreateHandler loops through the connection managers to create the handler h
func (dn *defaultNotifier) CreateHandler(h v1.Handler, c net.Conn) {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.CreateHandler(h, c)
	}
}

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

// TODO: The three methods below are currently NOT on the EventNotifier
//   interface, but are on the ConnectionManager interface -- should they
//   be on EventNotifier?

// DestroyHandler loops through the connection managers to destroy the handler h
// TODO: This name would need to change if we keep this
func (dn *defaultNotifier) DestroyHandler(h v1.Handler) {
	for _, connectionManager := range dn.connectionManagers {
		connectionManager.DestroyHandler(h)
	}
}

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
