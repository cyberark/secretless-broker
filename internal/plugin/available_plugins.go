package plugin

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// CompatiblePluginAPIVersion indicates what matching API version an external plugin
// must have so that we are capable of loading it.
var CompatiblePluginAPIVersion = "0.1.0"

// AvailablePlugins is an interface that provides a list of all the available
// plugins for each type that the broker supports.
type AvailablePlugins interface {
	HTTPPlugins() map[string]http.Plugin
	TCPPlugins() map[string]tcp.Plugin
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
) (AvailablePlugins, error) {

	return AllAvailablePluginsWithOptions(
		pluginDir,
		checksumsFile,
		GetInternalPluginsFunc,
		ExternalPlugins,
		logger,
	)
}

// AllAvailablePluginsWithOptions returns the full list of internal and external plugins
// available to the broker using explicitly-defined lookup functions.
func AllAvailablePluginsWithOptions(
	pluginDir string,
	checksumsFile string,
	internalLookupFunc InternalPluginLookupFunc,
	externalLookupfunc ExternalPluginLookupFunc,
	logger log.Logger,
) (AvailablePlugins, error) {

	internalPlugins, err := InternalPlugins(internalLookupFunc)
	if err != nil {
		return nil, err
	}

	externalPlugins, err := externalLookupfunc(
		pluginDir,
		checksumsFile,
		logger,
	)
	if err != nil {
		return nil, err
	}

	httpPlugins := map[string]http.Plugin{}

	for name, httpPlugin := range internalPlugins.HTTPPlugins() {
		if _, ok := httpPlugins[name]; ok {
			logger.Warnf("Internal plugin '%s' is replaced by an externally-provided plugin",
				name)
		}

		httpPlugins[name] = httpPlugin
	}
	for name, httpPlugin := range externalPlugins.HTTPPlugins() {
		if _, ok := httpPlugins[name]; ok {
			logger.Warnf("Internal plugin '%s' is replaced by an externally-provided plugin",
				name)
		}

		httpPlugins[name] = httpPlugin
	}

	tcpPlugins := map[string]tcp.Plugin{}

	for name, tcpPlugin := range internalPlugins.TCPPlugins() {
		if _, ok := tcpPlugins[name]; ok {
			logger.Warnf("Internal plugin '%s' is replaced by an externally-provided plugin",
				name)
		}

		tcpPlugins[name] = tcpPlugin
	}
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
