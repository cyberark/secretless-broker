package v1

import (
	"net"

	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// ConnectionManager is an interface to be implemented by plugins that want to
// manage connections for handlers and listeners.
type ConnectionManager interface {
	// Initialize is called before proxy initialization
	Initialize(config_v2.Config, func(config_v2.Config) error) error

	// CloseConnect is called when a client connection is closed
	CloseConnection(net.Conn)

	// ResolveCredential is called when a provider resolves a variable
	ResolveCredential(provider Provider, id string, value []byte)

	// ClientData is called for each inbound packet from clients
	ClientData(net.Conn, []byte)

	// ServerData is called for each inbound packet from the backend
	ServerData(net.Conn, []byte)

	// Shutdown is called when secretless caught a signal to exit
	Shutdown()
}
