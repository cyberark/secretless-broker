package main

// TODO: change the package name to your plugin name if this will be an internal connector

import (
	"net"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

/*
	NewConnector returns a tcp.Connector which returns an authenticated connection to a target service for each incoming
	client connection. It is a required method on the tcp.Plugin interface. The single argument passed in is of type
	connector.Resources. It contains connector-specific config and a logger.
*/
func NewConnector(conRes connector.Resources) tcp.Connector {
	connectorFunc := func(
		clientConn net.Conn,
		credentialValuesByID connector.CredentialValuesByID,
	) (backendConn net.Conn, err error) {
		// singleUseConnector is responsible for generating the authenticated connection
		// to the target service for each incoming client connection
		singleUseConnector := &SingleUseConnector{
			logger: conRes.Logger(),
			config: conRes.Config(), // Note: you may skip sending this if your plugin doesn't use any custom config
		}

		return singleUseConnector.Connect(clientConn, credentialValuesByID)
	}

	return tcp.ConnectorFunc(connectorFunc)
}

// PluginInfo is required as part of the Secretless plugin spec. It provides important metadata about the plugin.
func PluginInfo() map[string]string {
	// TODO: fill in the map according to
	// https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md#plugininfo
	return map[string]string{
		"pluginAPIVersion": "",
		"type":             "connector.tcp",
		"id":               "",
		"description":      "",
	}
}

// GetTCPPlugin is required as part of the Secretless plugin spec for TCP connector plugins. It returns the TCP plugin.
func GetTCPPlugin() tcp.Plugin {
	return tcp.ConnectorConstructor(NewConnector)
}
