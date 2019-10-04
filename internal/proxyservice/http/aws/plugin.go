package aws

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
)

// PluginInfo is required as part of the Secretless plugin spec. It provides
// important metadata about the plugin.
func PluginInfo() map[string]string {
	return map[string]string{
		"pluginAPIVersion": "0.1.0",
		"type":             "connector.http",
		"id":               "aws",
		"description":      "injects an HTTP request with AWS authorization headers",
	}
}

// NewConnector returns an http.Connector that decorates each incoming http
// request with authorization data.
//
// It is a required method on the http.Plugin interface. The single argument
// passed in is of type connector.Resources. It contains connector-specific
// config and a logger.
func NewConnector(conRes connector.Resources) http.Connector {
	return (&Connector{
		logger:   conRes.Logger(),
	}).Connect
}

// GetHTTPPlugin is required as part of the Secretless plugin spec for HTTP
// connector plugins. It returns the HTTP plugin.
func GetHTTPPlugin() http.Plugin {
	return http.ConnectorConstructor(NewConnector)
}
