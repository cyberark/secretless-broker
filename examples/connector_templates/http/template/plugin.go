package main

// TODO: change the package name to your plugin name if this will be an internal connector

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
)

// PluginInfo is required as part of the Secretless plugin spec. It provides important metadata about the plugin.
func PluginInfo() map[string]string {
	// TODO: fill in the map according to
	// https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md#plugininfo
	return map[string]string{
		"pluginAPIVersion": "",
		"type":             "connector.http",
		"id":               "",
		"description":      "",
	}
}

// NewConnector returns an http.Connector that decorates each incoming HTTP request
// so that it contains the necessary authentication information (typically by adding appropriate headers)
func NewConnector(conRes connector.Resources) http.Connector {
	return &Connector{
		logger: conRes.Logger(),
		config: conRes.Config(), // Note: you may skip sending this if your plugin doesn't use any custom config
	}
}

// GetHTTPPlugin is required as part of the Secretless plugin spec for HTTP connector plugins. It returns the HTTP plugin.
func GetHTTPPlugin() http.Plugin {
	return http.NewConnectorFunc(NewConnector)
}
