/*
Plugins are ways to add extended functionality to Secretless and the current interface version is v1.

WARNING: Given the speed of development, there will likely be cases of outdated documentation so please use this document
as a reference point and use the source code in this folder as the true representation of the API state!

Supported plugin types:
  -Listeners
  -Handlers
  -Connection managers

There is also an additional EventNotifier class used to bubble up events from listeners and handlers up
to the plugin manager but this class may be removed as we move more of the abstract functionality to the plugin manager
itself.

Basic overview

All plugins are currently loaded in the following manner:
  - Directory in `/usr/local/lib/secretless` is listed and any `*.so` files are iterated over. Sub-directory traversal is not supported at this time.
  - Each shared library plugin is searched for these variables:
    - PluginAPIVersion
    - PluginInfo
    - GetListeners
    - GetHandlers
    - GetConnectionManagers
  - Handlers are added to handler factory map.
  - Listeners are added to listener factory map.
  - Managers are added to manager factory map.
  - Managers are instantiated.
  - Listeners and handlers are instantiated by id whenever a configuration references them.


PluginAPIVersion

`PluginAPIVersion` (returns string) indicates the target API version of Secretless and must match the
supported version found at https://github.com/conjurinc/secretless/blob/master/internal/pkg/plugin/manager.go#L108 list in the
main daemon.

PluginInfo

This `string->string` map (returns `map[string]string`) has information about the plugin that the daemon might use for logging, prioritization, and masking.
While extraneous keys in the map are ignored, the map _must_ contain the following keys:

  - `version`
  Indicates the plugin version
  - `id`
  A computer-friendly id of the plugin. Naming should be constrained to short, spaceless ASCII lowercase alphanumeric set with a limited set of special characters (`-`, `_`, and `/`).
  - `name`
  User-friendly name of the plugin. This name will be used in most user-facing messages about the plugin and should be constrained in length to <30 chars.
  - `description`
  A longer description of the plugin though it should not exceed 100 characters.

GetListeners

Returns a map of provided listener ids to their factory methods (`map[string]func(v1.ListenerOptions) v1.Listener`) that
accept `v1.ListenerOptions` when invoked and return a new `v1.Listener` [listener](#listeners).

GetHandlers

Returns a map of provided handler ids to their factory methods (`map[string]func(v1.HandlerOptions) v1.Handler`) that
accept `v1.HandlerOptions` when invoked and return a new `v1.Handler` [handler](#handlers).

GetConnectionManagers


Returns a map of provided manager ids to their factory methods (`map[string]func() v1.ConnectionManager`) that
return a new `v1.ConnectionManager` connection manager when invoked.

Note: There is a high likelihood that this method will also have `v1.ConnectionManagerOptions` as the
factory parameter like the rest of the factory maps in the future releases

Example plugin

The following shows a sample plugin that conforms to the expected API:

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
*/
package v1
