package tcp

import (
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Plugin is the interface that TCP plugins need to implement.  Conceptually, a
// Plugin is is something that can create a Connector, and a Connector is
// something that knows how to "Connect", ie, how to authenticate tcp requests.
type Plugin interface {
	// NewConnector creates a new Connector based on the ConnectorResources
	// passed into it.
	NewConnector(connector.Resources) Connector
}

// Connector is an interface with a single method, "Connect".  "Connect" is a
// function that takes credentials and an unauthenticated connection and returns
// an authenticated connection to the target service.
type Connector interface {
	Connect(
		clientConn net.Conn,
		credentialValuesByID connector.CredentialValuesByID,
	) (backendConn net.Conn, err error)
}

// ConnectorFunc is a type that allows a free-standing "Connect" function to
// fulfill the Connector interface.  You simply cast a "Connect" function into
// a "ConnectorFunc", and it becomes a proper type with a "Connect" method,
// which simply calls the function itself.
type ConnectorFunc func(net.Conn, connector.CredentialValuesByID) (net.Conn, error)

// Connect is a ConnectorFunc's implementation of the Connector interface. It
// simply delegates to the underlying function itself.
func (cf ConnectorFunc) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (backendConn net.Conn, err error) {
	return cf(clientConn, credentialValuesByID)
}

// ConnectorConstructor allows a free-standing constructor function -- anything
// that takes connector.Resources and returns a Connector -- to fulfill the
// Plugin interface. Thus you can cast to it:
//
//     ConnectorConstructor(NewConnector)
//
// and you now have a Plugin. It does this by giving any function of the type
// ConnectorConstructor a constructor method called NewConnector that simply
// calls the function itself
type ConnectorConstructor func (connector.Resources) Connector

// NewConnector is a ConnectorConstructor's implementation of the Plugin
// interface. It simply delegates to the underlying function.
func (cc ConnectorConstructor) NewConnector(cr connector.Resources) Connector {
	return cc(cr)
}
