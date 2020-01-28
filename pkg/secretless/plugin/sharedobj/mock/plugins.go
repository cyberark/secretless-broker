package mock

import (
	"fmt"
	go_plugin "plugin"

	"github.com/cyberark/secretless-broker/internal/log"
	log_api "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

const pluginAPIVersion = "0.1.0"

// Plugins may appear on first glance to be duplication, but it's not. We
// can't use the actual implementation sharedobj.Plugin without creating a
// literal Go circular dependency.  It would be circular logically too: to use
// the thing we're testing to create a mock to test it.  Also, note it's purely
// coincidental that this implementation is the same as the actual
// implementation in sharedobj. Either one could change.  We only care about
// fulfilling the interface.
type Plugins struct {
	HTTPPluginsByID map[string]http.Plugin
	TCPPluginsByID  map[string]tcp.Plugin
}

// HTTPPlugins returns the mock HTTP plugins.
func (plugins *Plugins) HTTPPlugins() map[string]http.Plugin {
	return plugins.HTTPPluginsByID
}

// TCPPlugins returns the mock TCP plugins.
func (plugins *Plugins) TCPPlugins() map[string]tcp.Plugin {
	return plugins.TCPPluginsByID
}

// HTTPInternalPluginsByID returns mock HTTP plugins with internal ids.
func HTTPInternalPluginsByID() map[string]http.Plugin {
	return map[string]http.Plugin{
		"intHTTP1": NewHTTPPlugin("intHTTP1"),
		"intHTTP2": NewHTTPPlugin("intHTTP2"),
		"intHTTP3": NewHTTPPlugin("intHTTP3"),
	}
}

// HTTPExternalPluginsByID returns mock HTTP plugins with external ids.
func HTTPExternalPluginsByID() map[string]http.Plugin {
	return map[string]http.Plugin{
		"extHTTP1": NewHTTPPlugin("extHTTP1"),
		"extHTTP2": NewHTTPPlugin("extHTTP2"),
	}
}

// TCPInternalPluginsByID returns mock TCP plugins with internal ids.
func TCPInternalPluginsByID() map[string]tcp.Plugin {
	return map[string]tcp.Plugin{
		"intTCP1": NewTCPPlugin("intTCP1"),
		"intTCP2": NewTCPPlugin("intTCP2"),
		"intTCP3": NewTCPPlugin("intTCP3"),
	}
}

// TCPExternalPluginsByID returns mock TCP plugins with external ids.
func TCPExternalPluginsByID() map[string]tcp.Plugin {
	return map[string]tcp.Plugin{
		"extTCP1": NewTCPPlugin("extTCP1"),
		"extTCP2": NewTCPPlugin("extTCP2"),
		"extTCP3": NewTCPPlugin("extTCP3"),
	}
}

// InternalPlugins creates a mock AvailablePlugins for internal plugins.
func InternalPlugins() plugin.AvailablePlugins {
	return &Plugins{
		HTTPPluginsByID: HTTPInternalPluginsByID(),
		TCPPluginsByID:  TCPInternalPluginsByID(),
	}
}

// ExternalPlugins creates a mock AvailablePlugins for external plugins.
func ExternalPlugins() plugin.AvailablePlugins {
	return &Plugins{
		HTTPPluginsByID: HTTPExternalPluginsByID(),
		TCPPluginsByID:  TCPExternalPluginsByID(),
	}
}

// AllHTTPPlugins returns map combining the HTTP internal and external mock
// plugins.
func AllHTTPPlugins() map[string]http.Plugin {
	combined := InternalPlugins().HTTPPlugins()
	for k, v := range ExternalPlugins().HTTPPlugins() {
		combined[k] = v
	}
	return combined
}

// AllTCPPlugins returns map combining the TCP internal and external mock
// plugins.
func AllTCPPlugins() map[string]tcp.Plugin {
	combined := InternalPlugins().TCPPlugins()
	for k, v := range ExternalPlugins().TCPPlugins() {
		combined[k] = v
	}
	return combined
}

// GetInternalPlugins is function that returns the mock internal plugins.  It's
// needed to satisfy arguments of type InternalPluginLookupFunc.
func GetInternalPlugins() (plugin.AvailablePlugins, error) {
	return InternalPlugins(), nil
}

// GetExternalPlugins is function that returns the mock external plugins.  It's
// needed to satisfy arguments of type ExternalPluginLookupFunc.
func GetExternalPlugins(
	pluginDir string,
	checksumfile string,
	logger log_api.Logger,
) (plugin.AvailablePlugins, error) {
	return ExternalPlugins(), nil
}

// NewLogger returns a mock Logger.
func NewLogger() log_api.Logger {
	return log.New(true)
}

// RawPlugin implements the rawPlugin interface
type RawPlugin struct {
	PluginAPIVersion string
	PluginType       string
	PluginID         string
}

func (r RawPlugin) pluginInfo() map[string]string {
	return map[string]string{
		"pluginAPIVersion": r.PluginAPIVersion,
		"type":             r.PluginType,
		"id":               r.PluginID,
	}
}

func (r RawPlugin) httpPlugin() http.Plugin {
	return NewHTTPPlugin(r.PluginID)
}

func (r RawPlugin) tcpPlugin() tcp.Plugin {
	return NewTCPPlugin(r.PluginID)
}

// Lookup returns a go_plugin.Symbol for a given symbol name string. A go_plugin.Symbol
// is an empty interface, and can be instantiated with a function or a struct.
// This method is intended to mimic the Lookup method for a standard Go plugin.
func (r RawPlugin) Lookup(symbolName string) (go_plugin.Symbol, error) {
	switch symbolName {
	case "PluginInfo":
		return r.pluginInfo, nil
	case "GetHTTPPlugin":
		return r.httpPlugin, nil
	case "GetTCPPlugin":
		return r.tcpPlugin, nil
	}
	return nil, fmt.Errorf("unknown symbolName %s", symbolName)
}

// RawPlugins instantiates a map of RawPlugins for testing. Values of entries
// in this map should reflect values returned by HTTPExternalPluginsByID()
// and TCPExternalPluginsByID() above.
var RawPlugins = map[string]RawPlugin{
	"http1": RawPlugin{
		PluginAPIVersion: pluginAPIVersion,
		PluginType:       "connector.http",
		PluginID:         "extHTTP1",
	},
	"http2": RawPlugin{
		PluginAPIVersion: pluginAPIVersion,
		PluginType:       "connector.http",
		PluginID:         "extHTTP2",
	},
	"tcp1": RawPlugin{
		PluginAPIVersion: pluginAPIVersion,
		PluginType:       "connector.tcp",
		PluginID:         "extTCP1",
	},
	"tcp2": RawPlugin{
		PluginAPIVersion: pluginAPIVersion,
		PluginType:       "connector.tcp",
		PluginID:         "extTCP2",
	},
	"tcp3": RawPlugin{
		PluginAPIVersion: pluginAPIVersion,
		PluginType:       "connector.tcp",
		PluginID:         "extTCP3",
	},
}
