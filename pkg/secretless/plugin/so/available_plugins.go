package so

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// CompatiblePluginAPIVersion indicates what matching API version an external plugin
// must have so that we are capable of loading it.
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
		LoadPluginsFromDir,
		logger,
	)
}

// AllAvailablePluginsWithOptions returns the full list of internal and external
// plugins available to the broker using explicitly-defined lookup functions.
// TODO: Test this
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

	externalPlugins, err := ExternalPlugins(
		pluginDir,
		externalLookupfunc,
		logger,
		checksumsFile,
	)
	if err != nil {
		return nil, err
	}

	httpPlugins := internalPlugins.HTTPPlugins()
	for name, httpPlugin := range externalPlugins.HTTPPlugins() {
		if _, ok := httpPlugins[name]; ok {
			logger.Warnf("Internal plugin '%s' is replaced by an externally-provided plugin",
				name)
		}

		httpPlugins[name] = httpPlugin
	}

	tcpPlugins := internalPlugins.TCPPlugins()
	for name, tcpPlugin := range externalPlugins.TCPPlugins() {
		if _, ok := tcpPlugins[name]; ok {
			logger.Warnf("Internal plugin '%s' is replaced by an externally-provided plugin",
				name)
		}

		tcpPlugins[name] = tcpPlugin
	}

	return &Plugins{
		HTTPPluginsByID: httpPlugins,
		TCPPluginsByID:  tcpPlugins,
	}, nil
}
