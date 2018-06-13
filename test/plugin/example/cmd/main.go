package main

import (
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
	"github.com/conjurinc/secretless/test/plugin/example"
)

// Exports
var PluginApiVersion = "0.0.3"

var PluginInfo = map[string]string{
	"version":     "0.0.3",
	"id":          "example-plugin",
	"name":        "Example Plugin",
	"description": "Example plugin to demonstrate plugin functionality",
}

func GetListeners() map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener {
	return map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener{
		"echo": example.ListenerFactory,
	}
}

func GetHandlers() map[string]func() plugin_v1.Handler {
	return make(map[string]func() plugin_v1.Handler)
}

func GetManagers() map[string]plugin_v1.ConnectionManager {
	return map[string]plugin_v1.ConnectionManager{
		"example-plugin-manager": &example.ExampleManager{},
	}
}
