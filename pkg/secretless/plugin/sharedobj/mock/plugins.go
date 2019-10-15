package mock

import (
	"github.com/cyberark/secretless-broker/internal/log"
	log_api "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

var HTTPPluginsById = map[string]http.Plugin{
	"one": &HTTPPlugin{},
	"two": &HTTPPlugin{},
}

var TCPPluginsById = map[string]tcp.Plugin{
	"one":   &TCPPlugin{},
	"two":   &TCPPlugin{},
	"three": &TCPPlugin{},
}

// mockPlugins may appear on first glance to be duplication, but it's not. We
// can't use the actual implementation sharedobj.Plugin without creating a
// literal Go circular dependency.  It would be circular logically too: to use
// the thing we're testing to create a mock to test it.  Also, note it's purely
// coincidental that this implementation is the same as the actual
// implementation in sharedobj. Either one could change.  We only care about
// fulfilling the interface.
type mockPlugins struct {
	HTTPPluginsByID map[string]http.Plugin
	TCPPluginsByID  map[string]tcp.Plugin
}

func (plugins *mockPlugins) HTTPPlugins() map[string]http.Plugin {
	return plugins.HTTPPluginsByID
}

func (plugins *mockPlugins) TCPPlugins() map[string]tcp.Plugin {
	return plugins.TCPPluginsByID
}

// InternalPlugins creates an AvailablePlugins object composed of mocked plugins
func InternalPlugins() plugin.AvailablePlugins {
	return &mockPlugins{
		HTTPPluginsByID: map[string]http.Plugin{
			"intHTTP1": &HTTPPlugin{},
			"intHTTP2": &HTTPPlugin{},
			"intHTTP3": &HTTPPlugin{},
		},
		TCPPluginsByID: map[string]tcp.Plugin{
			"intTCP1": &TCPPlugin{},
			"intTCP2": &TCPPlugin{},
			"intTCP3": &TCPPlugin{},
		},
	}
}
func ExternalPlugins() plugin.AvailablePlugins {
	return &mockPlugins{
		HTTPPluginsByID: map[string]http.Plugin{
			"extHTTP1": &HTTPPlugin{},
			"extHTTP2": &HTTPPlugin{},
		},
		TCPPluginsByID: map[string]tcp.Plugin{
			"extTCP1": &TCPPlugin{},
			"extTCP2": &TCPPlugin{},
			"extTCP3": &TCPPlugin{},
		},
	}
}

func AllHTTPPlugins() map[string]http.Plugin {
	combined := InternalPlugins().HTTPPlugins()
	for k, v := range ExternalPlugins().HTTPPlugins() {
		combined[k] = v
	}
	return combined
}

func AllTCPPlugins() map[string]tcp.Plugin {
	combined := InternalPlugins().TCPPlugins()
	for k, v := range ExternalPlugins().TCPPlugins() {
		combined[k] = v
	}
	return combined
}

func GetInternalPlugins() (plugin.AvailablePlugins, error) {
	return InternalPlugins(), nil
}

func GetExternalPlugins(
	pluginDir string,
	checksumfile string,
	logger log_api.Logger,
) (plugin.AvailablePlugins, error) {
	return ExternalPlugins(), nil
}

func NewLogger() log_api.Logger {
	return log.New(true)
}
