package generic

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
		"id":               "generic_http",
		"description":      "Injects an HTTP request configurable headers.",
	}
}

// NewConnector returns an http.Connector that decorates each incoming http
// request with a basic auth header.
func NewConnector(conRes connector.Resources) http.Connector {
	logger := conRes.Logger()

	cfgYAML, err := NewConfigYAML(conRes.Config())
	if err != nil {
		logger.Panicf("can't create connector: can't unmarshal YAML.")
	}

	cfg, err := newConfig(cfgYAML)
	if err != nil {
		logger.Panicf("can't create connector: can't validate YAML.")
	}

	return &Connector{
		logger: logger,
		config: cfg,
	}
}

// GetHTTPPlugin is required as part of the Secretless plugin spec for HTTP
// connector plugins. It returns the HTTP plugin.
func GetHTTPPlugin() http.Plugin {
	return http.NewConnectorFunc(NewConnector)
}
