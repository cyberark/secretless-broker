package sharedobj

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// CompatiblePluginAPIVersion indicates what matching API version an external plugin
// must have so that Secretless is capable of loading it.
var CompatiblePluginAPIVersion = "0.1.0"

// IsHTTPPlugin uses AvailablePlugins to determine if a pluginId is an HTTP
// plugin.
func IsHTTPPlugin(availPlugins plugin.AvailablePlugins, pluginID string) bool {
	for id := range availPlugins.HTTPPlugins() {
		if pluginID == id {
			return true
		}
	}
	return false
}

// AllAvailablePlugins returns the full list of internal and external plugins
// available to the broker.
func AllAvailablePlugins(
	pluginDir string,
	checksumsFile string,
	logger log.Logger,
) (plugin.AvailablePlugins, error) {

	return AllAvailablePluginsWithOptions(
		pluginDir,
		checksumsFile,
		GetInternalPluginsFunc,
		ExternalPlugins,
		logger,
	)
}

// AllAvailablePluginsWithOptions returns the full list of internal and external
// plugins available to the broker using explicitly-defined lookup functions.
func AllAvailablePluginsWithOptions(
	pluginDir string,
	checksumsFile string,
	internalLookupFunc InternalPluginLookupFunc,
	externalLookupfunc ExternalPluginLookupFunc,
	logger log.Logger,
) (plugin.AvailablePlugins, error) {

	internalPlugins, err := InternalPlugins(internalLookupFunc)
	if err != nil {
		return nil, err
	}

	externalPlugins, err := externalLookupfunc(pluginDir, checksumsFile, logger)
	if err != nil {
		return nil, err
	}

	httpPlugins := map[string]http.Plugin{}

	for name, httpPlugin := range internalPlugins.HTTPPlugins() {
		if _, ok := httpPlugins[name]; ok {
			// TODO: Should this ever happen?  Do we need this check?  Should it panic?
			logger.Warnf("Internal plugin '%s' replaced by internal plugin", name)
		}
		httpPlugins[name] = httpPlugin
	}

	for name, httpPlugin := range externalPlugins.HTTPPlugins() {
		if _, ok := httpPlugins[name]; ok {
			logger.Warnf("Internal plugin '%s' replaced by external plugin", name)
		}
		httpPlugins[name] = httpPlugin
	}

	tcpPlugins := map[string]tcp.Plugin{}

	for name, tcpPlugin := range internalPlugins.TCPPlugins() {
		if _, ok := tcpPlugins[name]; ok {
			logger.Warnf("Internal plugin '%s' replaced by internal plugin", name)
		}
		tcpPlugins[name] = tcpPlugin
	}

	for name, tcpPlugin := range externalPlugins.TCPPlugins() {
		if _, ok := tcpPlugins[name]; ok {
			logger.Warnf("Internal plugin '%s' replaced by external plugin", name)
		}
		tcpPlugins[name] = tcpPlugin
	}

	return &Plugins{
		HTTPPluginsByID: httpPlugins,
		TCPPluginsByID:  tcpPlugins,
	}, nil
}

// NewPlugins plugins creates a new instance of Plugins with both maps
// initialized but empty.
func NewPlugins() Plugins {
	return Plugins{
		HTTPPluginsByID: map[string]http.Plugin{},
		TCPPluginsByID:  map[string]tcp.Plugin{},
	}
}
// Plugins represent a holding object for a bundle of plugins of different types.
type Plugins struct {
	HTTPPluginsByID map[string]http.Plugin
	TCPPluginsByID  map[string]tcp.Plugin
}

// HTTPPlugins returns only the HTTP plugins in the Plugins struct.
func (plugins *Plugins) HTTPPlugins() map[string]http.Plugin {
	return plugins.HTTPPluginsByID
}

// TCPPlugins returns only the TCP plugins in the Plugins struct.
func (plugins *Plugins) TCPPlugins() map[string]tcp.Plugin {
	return plugins.TCPPluginsByID
}
