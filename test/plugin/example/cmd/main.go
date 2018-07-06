package main

import (
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
	"github.com/conjurinc/secretless/test/plugin/example"
)

// Exports
var PluginApiVersion = "0.0.6"

var PluginInfo = map[string]string{
	"version":     "0.0.6",
	"id":          "example-plugin",
	"name":        "Example Plugin",
	"description": "Example plugin to demonstrate plugin functionality",
}

func GetListeners() map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener {
	return map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener{
		"echo": example.ListenerFactory,
	}
}

func GetHandlers() map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler {
	return map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler{
		"example-handler": example.HandlerFactory,
	}
}

func GetConnectionManagers() map[string]func() plugin_v1.ConnectionManager {
	return map[string]func() plugin_v1.ConnectionManager{
		"example-plugin-manager": example.ManagerFactory,
	}
}
