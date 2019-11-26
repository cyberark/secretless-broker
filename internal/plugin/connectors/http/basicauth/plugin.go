package basicauth

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
		"id":               "basic_auth",
		"description":      "injects an HTTP request with a Basic auth header",
	}
}

// GetHTTPPlugin is required as part of the Secretless plugin spec for HTTP
// connector plugins. It returns the HTTP plugin.
func GetHTTPPlugin() http.Plugin {
	newConnector, err := generic.NewConnectorConstructor(
		&generic.ConfigYAML{
			CredentialValidations: map[string]string{
				"username": "[^:]+",
			},
			Headers: map[string]string{
				"Authorization": "Basic {{ printf \"%s:%s\" .username .password | base64 }}",
			},
		},
	)

	// This should never occur at runtime.  The only way it could is if the
	// ConfigYAML definition above was faulty.  And if this were the case, our
	// tests would be broken.
	if err != nil {
		log.Panicf("Failed to create generic HTTP NewConnector: %s", err)
	}
	return newConnector
}
