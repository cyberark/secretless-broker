package tcp

import (
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Plugin is the interface that TCP plugins need to implement.  Conceptually, a
// Plugin is is something that can create a Connector, and a Connector is
// something that knows how to "Connect", i.e., how to authenticate TCP requests.
type Plugin interface {
	// NewConnector creates a new Connector based on the ConnectorResources
	// passed into it.
	NewConnector(connector.Resources) Connector
}

/*
Connector is an interface with a single method, Connect. Connect is a
function that takes credentials and an unauthenticated connection and returns
an authenticated connection to the target service.

The authentication stage is complete after Connect is called. At that point,
Secretless has both the client connection and an authenticated
connection to the target service.  The relationship between the client
connection, Secretless, and the authenticated target service (backend) connection looks
like this:

	clientConn <--> Secretless <--> backendConn

Once the authentication stage is complete, Secretless becomes an invisible proxy,
streaming bytes back and forth between client and target service as if they were directly
connected.
*/
type Connector interface {
	Connect(
		clientConn net.Conn,
		credentialValuesByID connector.CredentialValuesByID,
	) (backendConn net.Conn, err error)
}

// ConnectorFunc is a type that allows a free-standing "Connect" function to
// fulfill the Connector interface.  You cast a "Connect" function into
// a "ConnectorFunc", and it becomes a proper type with a "Connect" method
// that calls the function itself.
type ConnectorFunc func(net.Conn, connector.CredentialValuesByID) (net.Conn, error)

// Connect is a ConnectorFunc's implementation of the Connector interface. It
// delegates to the underlying function itself.
func (cf ConnectorFunc) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (backendConn net.Conn, err error) {
	return cf(clientConn, credentialValuesByID)
}

/*
ConnectorConstructor makes it easy to define Secretless TCP connector plugins.
If your plugin definition includes a standalone constructor function that takes a
connector.Resources as input and returns a Connector, you can cast your function
as a ConnectorConstructor, which is a type that already fulfills the plugin interface.

This is possible because the ConnectorConstructor function type has a constructor
method called  NewConnector that simply calls the function itself.

For example, in your plugin definition you can define the NewConnector
constructor for your plugin and call:

	func GetTCPPlugin() tcp.Plugin {
		return tcp.ConnectorConstructor(NewConnector)
	}

to fulfill the Plugin interface.
*/
type ConnectorConstructor func(connector.Resources) Connector

/*
NewConnector returns a Connector, a one method interface that performs the actual
authentication.

When Secretless runs, it calls NewConnector once and then holds onto the
returned Connector.  That Connector (remember: it's just a a single method)
is then called each time a new client connection requires authentication.
*/
func (cc ConnectorConstructor) NewConnector(cr connector.Resources) Connector {
	return cc(cr)
}
