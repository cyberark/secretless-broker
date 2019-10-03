package mysql

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// NewConnector is a required method on the tcp.Plugin interface. It returns a
// tcp.Connector.
//
// The single argument passed in is of type connector.Resources. It contains
// connector-specific config and a logger.
func NewConnector(conRes connector.Resources) tcp.Connector {
	return (&Connector{
		logger:   conRes.Logger(),
	}).Connect
}

// PluginInfo is required as part of the Secretless plugin spec. It provides
// important metadata about the plugin.
func PluginInfo() map[string]string {
	return map[string]string{
		"pluginAPIVersion": "0.1.0",
		"type":             "connector.tcp",
		"id":               "mysql",
		"description":      "returns an authenticated connection to a MySQL database",
	}
}

// GetTCPPlugin is required as part of the Secretless plugin spec for TCP connector
// plugins. It returns the TCP plugin.
func GetTCPPlugin() tcp.Plugin {
	return tcp.ConnectorConstructor(NewConnector)
}
