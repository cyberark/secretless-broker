package tcp

import (
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Plugin is the main interface that TCP plugins need to implement
// to be loaded by our codebase.
type Plugin interface {
	// NewConnector creates a new Connector based on the ConnectorResources
	// passed into it.
	NewConnector(connector.Resources) Connector
}

// Connector is the function that will be invoked when a matching
// TCP request comes in. It uses both the initiating connection and the
// secrets map to authenticate the client, returning the backend
// network connection.
type Connector func(
	clientConn net.Conn,
	secrets connector.SecretsByID,
) (backendConn net.Conn, err error)
