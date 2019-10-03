package http

import (
	"net/http"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Plugin is the main interface that HTTP plugins need to implement
// to be loaded by our codebase.
type Plugin interface {
	// NewConnector creates a new Connector based on the ConnectorResources
	// passed into it.
	NewConnector(connector.Resources) Connector
}

// Connector is the function that will be invoked when a matching
// request comes in. It uses both the request object and the credentials
// map to authenticate the client.
type Connector func(
	request *http.Request,
	credentialValuesByID connector.CredentialValuesByID,
) error

// ConnectorConstructor, through type-conversion e.g. ConnectorConstructor(NewConnector),
// allows a free-standing NewConnector func to conform to the http.Plugin interface without
// the need for additional boilerplate. It does this by giving any function of the type
// ConnectorConstructor a constructor method called NewConnector that simply calls the function
// itself
type ConnectorConstructor func (connector.Resources) Connector
func (cc ConnectorConstructor) NewConnector(cr connector.Resources) Connector {
	return cc(cr)
}
