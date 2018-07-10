package main

import (
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
	"github.com/conjurinc/secretless/test/plugin/example"
)

// PluginAPIVersion is the API version being used
var PluginAPIVersion = "0.0.7"

// PluginInfo describes the plugin
var PluginInfo = map[string]string{
	"version":     "0.0.7",
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

// GetConnectionManagers returns the example connection manager
func GetConnectionManagers() map[string]func() plugin_v1.ConnectionManager {
	return map[string]func() plugin_v1.ConnectionManager{
		"example-plugin-manager": example.ManagerFactory,
	}
}
