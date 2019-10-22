# IMPORTANT: API OVERHAUL IN PROGRESS

**The plugin API is changing, and this version is now private.  The new public API is
available in `pkg/plugin`. For convenience, the old documentation is provided below.
However, no new plugins should be built on this framework.**

# OLD README CONTENT

# Plugin v1 API

**_WARNING: Given the speed of development, there will likely be cases of outdated documentation so please use this document
as a reference point and use the source code in this folder as the true representation of the API state!_**

- [Basic Overview](#basic-overview)
- [Supported Plugin Types](#supported-plugin-types)
  - [Listeners](#listeners)
  - [Handlers](#handlers)
  - [Connection managers](#connection-managers)
- [Example Plugin Skeleton](#example-plugin)

## Supported plugin types

- [Listeners](#listeners)
- [Handlers](#handlers)
- [Connection managers](#connection-managers)

There is also an additional [EventNotifier](#eventnotifier) class used to bubble up events from listeners and handlers up
to the plugin manager but this class may be removed as we move more of the abstract functionality to the plugin manager
itself.

## Basic overview

All plugins are currently loaded in the following manner:
- Directory in `/usr/local/lib/secretless` is listed and any `*.so` files are iterated over. Sub-directory traversal
is not supported at this time.
- Each shared library plugin is searched for these variables:
  - [`PluginAPIVersion`](#pluginapiversion)
  - [`PluginInfo`](#plugininfo)
  - [`GetListeners`](#getlisteners)
  - [`GetHandlers`](#gethandlers)
  - [`GetConnectionManagers`](#getconnectionmanagers)
 - Handlers are added to handler factory map.
 - Listeners are added to listener factory map.
 - Managers are added to manager factory map.
 - Managers are instantiated.
 - Listeners and handlers are instantiated by id whenever a configuration references them.

 ### PluginAPIVersion
 (returns `string`)

`PluginAPIVersion` string indicates the target API version of the Secretless Broker and must match the
[supported version](https://github.com/cyberark/secretless-broker/blob/master/internal/plugin/manager.go#L108) list in the
main daemon.

### PluginInfo
(returns `map[string]string`)

This `string->string` map has information about the plugin that the daemon might use for logging, prioritization, and masking.
While extraneous keys in the map are ignored, the map _must_ contain the following keys:
- `version`: Indicates the plugin version
- `id`: A computer-friendly id of the plugin. Naming should be constrained to short, spaceless ASCII lowercase alphanumeric
set with a limited set of special characters (`-`, `_`, and `/`).
- `name`: User-friendly name of the plugin. This name will be used in most user-facing messages about the plugin and should
be constrained in length to <30 chars.
- `description`: A longer description of the plugin though it should not exceed 100 characters.

### GetListeners
(returns `map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener`)

Returns a map of provided listener ids to their factory methods that accept `plugin_v1.ListenerOptions` when invoked and
return a new `plugin_v1.Listener` [listener](#listeners).

### GetHandlers
(returns `map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler`)

Returns a map of provided handler ids to their factory methods that accept `plugin_v1.HandlerOptions` when invoked and
return a new `plugin_v1.Handler` [handler](#handlers).

### GetConnectionManagers
(returns `map[string]func() plugin_v1.ConnectionManager`)

Returns a map of provided manager ids to their factory methods that return a new `plugin_v1.ConnectionManager`
[connection manager](#connection-managers) when invoked.

_Note: There is a high likelihood that this method will also have `plugin_v1.ConnectionManagerOptions` as the
factory parameter like the rest of the factory maps in the future releases_

## Example plugin
```
var PluginAPIVersion = "0.0.7"

var PluginInfo = map[string]string{
	"version":     "0.0.7",
	"id":          "test-plugin",
	"name":        "Test Plugin",
	"description": "Test plugin to demonstrate plugin functionality",
}

func GetListeners() map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener {
	return map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener{
		"test-listener-plugin": testPlugin.ListenerFactory,
	}
}

func GetHandlers() map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler {
	return map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler{
		"test-handler-plugin": testPlugin.HandlerFactory,
	}
}

func GetConnectionManagers() map[string]func() plugin_v1.ConnectionManager {
	return map[string]func() plugin_v1.ConnectionManager{
		"test-connection-manager-plugin": testPlugin.ConnectionManagerFactory,
	}
}
```

## Listeners

Listeners are generally an ingress IP (TCP or UDP) port and/or socket file listener that is the target for the
downstream client of a service. The listeners usually listens on the socket or port for inbound connections and
then spawns [Handlers](#handlers) for any new connection to them.

[Current API](https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin_v1/listener.go)

## Handlers

Handlers are connection state objects that get instantiated on each new connection to a listener that provide
connectivity between:

- Downstream to the proxy server
- Proxy server to upstream server

As part of this functionality, they also modify traffic at connection-level to provide the injection of credentials
for the particular type of protocol they are handling though majority of their functionality is in simple shuttling
of data between downstream and upstream in a transparent manner.

_Note: The handler API interface contains a few methods/fields that were unable to be abstracted away and provide
support for all the protocols at the time of writing this note and those methods and fields are likely to get
removed/changed in future versions of the Secretless Broker_

[Current API](https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin_v1/handler.go)

## Connection Managers

Connection managers are plugins that can be used to both monitor and control the Secretless Broker. They provide callbacks
for various events that are happening and can manage that information and act on it.

_Note: While the API interface is currently expressive enough to provide basic functionality for the intended
purpose, the eventing is still being worked on heavily and the APIs/eventing triggers are extremely likely to
change in the near future_

[Current API](https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin_v1/connection_manager.go)

## EventNotifier

`EventNotifier` is used as a target object of events for handlers and listeners that notifies the plugin manager
in an abstract way without needing to pass down the full connection manager as a parameter.

_Note: Currently not all included listeners and handlers use this eventing but full support for that is planned
in the future releases_

[Current API](https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin_v1/event_notifier.go)
