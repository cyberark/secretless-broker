package mysql

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// NewConnector is required method on the tcp.Plugin interface. It returns a
// tcp.Connector.
//
// The single argument passed in is of type connector.Resources. It contains
// connector-specific config and a logger.
func NewConnector(conRes connector.Resources) tcp.Connector {
	return (&Connector{
		logger:   conRes.Logger(),
	}).Connect
}

// pluginWrapper is a wrapper type that makes it possible for the NewConnector func
// to stand alone as a tcp.Plugin
type pluginWrapper func (connector.Resources) tcp.Connector
func (pw pluginWrapper) NewConnector(cr connector.Resources) tcp.Connector {
	return pw(cr)
}

// PluginInfo is required as part of the Secretless pluginWrapper spec. It provides
// important metadata about the pluginWrapper.
func PluginInfo() map[string]string {
	return map[string]string{
		"pluginAPIVersion": "0.1.0",
		"type":             "connector.tcp",
		"id":               "mysql",
		"description":      "returns an authenticated connection to a MySQL database",
	}
}

// GetTCPPlugin is required as part of the Secretless pluginWrapper spec for TCP connector
// plugins. It returns the TCP pluginWrapper.
func GetTCPPlugin() tcp.Plugin {
	return pluginWrapper(NewConnector)
}