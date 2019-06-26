package main

import (
	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/test/plugin/example"
)

// PluginAPIVersion is the API version being used
var PluginAPIVersion = "0.0.8"

// PluginInfo describes the plugin
var PluginInfo = map[string]string{
	"version":     "0.0.8",
	"id":          "example-plugin",
	"name":        "Example Plugin",
	"description": "Example plugin to demonstrate plugin functionality",
}

// GetListeners returns the echo listener
func GetListeners() map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener {
	return map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener{
		"echo": example.ListenerFactory,
	}
}

// GetHandlers returns the example handler
func GetHandlers() map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler {
	return map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler{
		"example-handler": example.HandlerFactory,
	}
}

// GetProviders returns the example provider
func GetProviders() map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	return map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error){
		"example-provider": example.ProviderFactory,
	}
}

// GetConnectionManagers returns the example connection manager
func GetConnectionManagers() map[string]func() plugin_v1.ConnectionManager {
	return map[string]func() plugin_v1.ConnectionManager{
		"example-plugin-connection-manager": example.ConnManagerFactory,
	}
}

// GetConfigurationManagers returns the example configuration manager
func GetConfigurationManagers() map[string]func(plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager {
	return map[string]func(plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager{
		"example-plugin-config-manager": example.ConfigManagerFactory,
	}
}
