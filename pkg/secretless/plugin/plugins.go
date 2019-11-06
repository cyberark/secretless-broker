package plugin

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// builtinConnectorIDs is a list of connector IDs for "connectors" that don't
// even exist as separate objects, but are hardcoded into the ssh proxy
// services.
var builtinConnectorIDs = []string{"ssh", "ssh-agent"}

// AvailablePlugins is an interface that provides a list of all the available
// plugins for each type that the broker supports.
type AvailablePlugins interface {
	HTTPPlugins() map[string]http.Plugin
	TCPPlugins() map[string]tcp.Plugin
}

// AvailableConnectorIDs returns a list of all available connector IDs: for both
// builtin connectors and those provided by AvailablePlugins. In the case of
// AvailablePlugins, the connector ID and plugin ID are identical.  For the
// builtin "connectors", the concept of plugin ID doesn't make sense.
// AvailableConnectorIDs is a pure function that depends only on the
// AvailablePlugins interface, which is why we define it here rather than in the
// implementation package "sharedobj".
func AvailableConnectorIDs(availPlugins AvailablePlugins) []string {
	// Start with the IDs of the static, built-in ssh plugins.
	var connectorIDs []string
	connectorIDs = append(connectorIDs, builtinConnectorIDs...)

	// Now add the connector IDs from the available plugins.
	for name := range availPlugins.TCPPlugins() {
		connectorIDs = append(connectorIDs, name)
	}
	for name := range availPlugins.HTTPPlugins() {
		connectorIDs = append(connectorIDs, name)
	}
	return connectorIDs
}
