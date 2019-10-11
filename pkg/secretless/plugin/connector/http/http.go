package http

import (
	"net/http"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Plugin is the interface that HTTP plugins need to implement.  Conceptually, a
// Plugin is is something that can create a Connector, and a Connector is
// something that knows how to "Connect", ie, how to authenticate http requests.
type Plugin interface {
	// NewConnector creates a new Connector based on the ConnectorResources
	// passed into it.
	NewConnector(connector.Resources) Connector
}

// Connector is an interface with a single method, "Connect".  "Connect" is a
// function that knows how to authenticate an http request. It uses
// its credential values argument passed in to modify the http.Request
// request object and the credentials map to authenticate the client.
type Connector interface {
	Connect(
		request *http.Request,
		credentialValuesByID connector.CredentialValuesByID,
	) error
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
