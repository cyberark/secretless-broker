package conjur

import (
	"log"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/http/generic"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
)

// PluginInfo is required as part of the Secretless plugin spec. It provides
// important metadata about the plugin.
func PluginInfo() map[string]string {
	return map[string]string{
		"pluginAPIVersion": "0.1.0",
		"type":             "connector.http",
		"id":               "conjur",
		"description":      "injects an HTTP request with Conjur's auth header",
	}
}

// GetHTTPPlugin is required as part of the Secretless plugin spec for HTTP
// connector plugins. It returns the HTTP plugin.
func GetHTTPPlugin() http.Plugin {
	newConnector, err := generic.NewConnectorConstructor(
		&generic.ConfigYAML{
			Headers: map[string]string{
				"Authorization": `Token token="{{ .accessToken | base64 }}"`,
			},
		},
	)

	// This error should never occur at runtime. It could only happen if the ConfigYAML
	// definition above was faulty. And - if that was the case - our tests would
	// also be broken.
	if err != nil {
		log.Panicf("Failed to create Conjur HTTP NewConnector: %s", err)
	}

	return newConnector
}
