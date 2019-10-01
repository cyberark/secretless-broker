package sharedobj

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// InternalPluginLookupFunc returns all available internal plugins.
type InternalPluginLookupFunc func() (plugin.AvailablePlugins, error)

// GetInternalPluginsFunc returns currently available internal plugins
// but for now, this list is empty since we have none implemented.
func GetInternalPluginsFunc() (plugin.AvailablePlugins, error) {
	return &Plugins{
		HTTPPluginsByID: map[string]http.Plugin{},
		TCPPluginsByID:  map[string]tcp.Plugin{},
	}, nil
}

// InternalPlugins is used to enumerate internally-available plugins to the clients
// of this method.
func InternalPlugins(lookupFunc InternalPluginLookupFunc) (plugin.AvailablePlugins, error) {
	plugins, err := lookupFunc()
	if err != nil {
		return nil, err
	}

	if plugins == nil {
		plugins = &Plugins{
			HTTPPluginsByID: nil,
			TCPPluginsByID:  nil,
		}
	}

	return plugins, nil
}
