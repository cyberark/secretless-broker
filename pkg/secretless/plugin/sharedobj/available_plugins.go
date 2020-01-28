package sharedobj

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

const pluginConflictMessage = "%s plugin ID '%s' conflicts with an existing internal plugin"

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

// checkPluginIDConflicts asserts that a given plugin ID is not used
// by any internal HTTP or TCP plugin.
func checkPluginIDConflicts(
	pluginType string, // "HTTP" or "TCP"
	pluginID string,
	internalPlugins plugin.AvailablePlugins,
	logger log.Logger) {

	httpPlugins := internalPlugins.HTTPPlugins()
	if _, ok := httpPlugins[pluginID]; ok {
		logger.Panicf(pluginConflictMessage, pluginType, pluginID)
	}
	tcpPlugins := internalPlugins.TCPPlugins()
	if _, ok := tcpPlugins[pluginID]; ok {
		logger.Panicf(pluginConflictMessage, pluginType, pluginID)
	}
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

	allHTTPPlugins := map[string]http.Plugin{}
	allTCPPlugins := map[string]tcp.Plugin{}

	// Assemble internal plugins. Plugin IDs for internal plugins are
	// assumed to be unique because their definitions are hardcoded.
	internalPlugins, err := InternalPlugins(internalLookupFunc)
	if err != nil {
		return nil, err
	}
	for pluginID, httpPlugin := range internalPlugins.HTTPPlugins() {
		allHTTPPlugins[pluginID] = httpPlugin
	}
	for pluginID, tcpPlugin := range internalPlugins.TCPPlugins() {
		allTCPPlugins[pluginID] = tcpPlugin
	}

	// Assemble external plugins. Check whether the plugin ID for each
	// external plugin conflicts with any plugin IDs of internal plugins.
	// (Checks for uniqueness among external HTTP and TCP plugins is
	// done elsewhere, i.e. as external plugins are discovered.)
	externalPlugins, err := externalLookupfunc(pluginDir, checksumsFile, logger)
	if err != nil {
		return nil, err
	}
	for pluginID, httpPlugin := range externalPlugins.HTTPPlugins() {
		checkPluginIDConflicts("HTTP", pluginID, internalPlugins, logger)
		allHTTPPlugins[pluginID] = httpPlugin
	}
	for pluginID, tcpPlugin := range externalPlugins.TCPPlugins() {
		checkPluginIDConflicts("TCP", pluginID, internalPlugins, logger)
		allTCPPlugins[pluginID] = tcpPlugin
	}

	return &Plugins{
		HTTPPluginsByID: allHTTPPlugins,
		TCPPluginsByID:  allTCPPlugins,
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
