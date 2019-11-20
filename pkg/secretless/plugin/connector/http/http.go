package http

import (
	"net/http"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Plugin is the interface that HTTP plugins need to implement.  Conceptually, a
// Plugin is something that can create a Connector, and a Connector is something
// that knows how to "Connect", i.e., how to authenticate HTTP requests.
type Plugin interface {
	// NewConnector creates a new Connector based on the ConnectorResources
	// passed into it.
	NewConnector(connector.Resources) Connector
}

/*
Connector is an interface with a single method, Connect. Connect is a
function that knows how to authenticate an HTTP request. It uses
its credential values argument to modify the http.Request and authenticate the
client.

Typically, altering the HTTP request so that it contains the necessary
authentication information means adding the appropriate headers to the request
-- for example, an Authorization header containing a Token, or a header
containing an API key.

Since HTTP is a stateless protocol, Secretless calls the Connect function every
time a client sends an HTTP request to the target server, so that every request
is authenticated.
*/
type Connector interface {
	Connect(
		request *http.Request,
		credentialValuesByID connector.CredentialValuesByID,
	) error
}

/*
NewConnectorFunc makes it easy to define Secretless HTTP connector plugins.
If your plugin definition includes a standalone constructor function that takes a
connector.Resources as input and returns a Connector, you can cast your function
as a NewConnectorFunc, which is a type that already fulfills the plugin interface.

This is possible because the NewConnectorFunc function type has a constructor
method called  NewConnector that simply calls the function itself.

For example, in your plugin definition you can define the NewConnector
constructor for your plugin and call:

	func GetHTTPPlugin() http.Plugin {
		return http.NewConnectorFunc(NewConnector)
	}

to fulfill the Plugin interface.
*/
type NewConnectorFunc func(connector.Resources) Connector

/*
NewConnector returns a Connector, which is a one method interface that performs the actual
authentication.

When Secretless runs, it calls NewConnector once and then holds onto the
returned Connector.  That Connector (remember: it's just a a single method)
is then called each time a new client connection requires authentication.
*/
func (cc NewConnectorFunc) NewConnector(cr connector.Resources) Connector {
	return cc(cr)
}
