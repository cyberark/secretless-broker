# Developer Map

This document is meant to provide a mapping from the conceptual model of Secretless to the codebase.

## Directories

+ [./cmd](./cmd) - entrypoints for [binaries](#binaries)
+ [./internal/app](./internal/app) - internal implementations of [plugins](#plugins)
+ [./internal/pkg](./internal/pkg) - [plugin manager](#pluginmanager) and [resolver](#resolver)
+ [./pkg/secretless](./pkg/secretless) - secretless [config](#config) and [plugin interfaces](#plugins)

## Binaries

The entrypoints to binaries exist in "./cmd/${binary_name}"

2 binaries are provided in the secretless repo.
+ [secretless](#secretless)
+ summon2

## Secretless

### Configuration
config.Config
[./pkg/secretless/config/config.go](./pkg/secretless/config/config.go)

+ represents the secretless configuration as struct

### Entrypoint
[./cmd/secretless-broker/main.go](./cmd/secretless-broker/main.go)
obtains the following via flags:

+ -f secretless config file
+ -p plugin directory
+ --watch condition to reload when config file changes
+ --debug toggles debug information
+ --config-mgr strategy to provide secretless configuration

plugins manager:
  + loads internal [plugins](#plugins)
  + loads external [plugins](#plugins) from the plugin directory
  + registers signal handlers
  + runs Proxy

### PluginManager 
plugin.Manager

[./internal/pkg/plugin/manager.go](./internal/pkg/plugin/manager.go)

+ is a singleton
+ loads and holds [plugins](#plugins)
+ creates
  + [Resolver](#resolver)
  + [Proxy](#proxy)
+ root of
  + RunHandlerFunc
  + RunListenerFunc
+ manages all the [plugins](#plugins) 
+ initialises and runs [Proxy](#proxy) 
+ implements [EventNotifier](#eventnotifier) (only one in the whole repo)


## Plugins
Plugins are used as a mechanism for extending Secretless functionality beyond the core. These are expressed as interfaces. They are versioned in anticipation of growth.

+ v1 interfaces are defined at [./pkg/secretless/plugin/v1](./pkg/secretless/plugin/v1)
+ list of interfaces
  + ConfigurationManager 
    - pushes configuration data and triggers updates
    - interface: [./pkg/secretless/plugin/v1/configuration_manager.go](./pkg/secretless/plugin/v1/configuration_manager.go)
    - implementation: [./internal/app/secretless/configurationmanagers/configfile](./internal/app/secretless/configurationmanagers/configfile)
  + ConnectionManager
    - manages connections for handlers and listeners
    - interface: ./pkg/secretless/plugin/v1/connection_manager.go
  + [EventNotifier](#eventnotifier)
    - passes events to plugin manager
    - interface: [./pkg/secretless/plugin/v1/event_notifier.go](./pkg/secretless/plugin/v1/event_notifier.go)
    - implementation: [./internal/pkg/plugin/manager.go](./internal/pkg/plugin/manager.go)
  + [Listener](#listener)
    - listens and accepts client connections and passes them to a handler
    - interface: [./pkg/secretless/plugin/v1/listener.go](./pkg/secretless/plugin/v1/listener.go)
    - pg implementation: [./internal/app/secretless/listeners/pg/listener.go](./internal/app/secretless/listeners/pg/listener.go)
  + [Handler](#handler)
    - receives client connection, connect it to a backend and streams connection.
    - interface: [./pkg/secretless/plugin/v1/handler.go](./pkg/secretless/plugin/v1/handler.go)
    - pg implementation: [./internal/app/secretless/handlers/pg/handler.go](./internal/app/secretless/handlers/pg/handler.go)
  + Provider
    - used to obtain values from a secret vault backend
    - interface: [./pkg/secretless/plugin/v1/provider.go](./pkg/secretless/plugin/v1/provider.go)
    - env implementation: [./internal/app/secretless/providers/env/provider.go](./internal/app/secretless/providers/env/provider.go)
  + [Resolver](#resolver)
    - manages Providers and provides convenient interface to obtain multiple values from multiple secret vault backends
    - interface: [./pkg/secretless/plugin/v1/resolver.go](./pkg/secretless/plugin/v1/resolver.go)
    - implementation: [./internal/pkg/plugin/resolver.go](./internal/pkg/plugin/resolver.go)
 
+ internal implementations of most plugin interfaces
  + located at ./internal/app/secretless
  + most plugins are exposed as factories, where the key is the identifier and the value is an interface pointer e.g. map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler

### EventNotifier
interface: [./pkg/secretless/plugin/v1/event_notifier.go](./pkg/secretless/plugin/v1/event_notifier.go)

implementation: [./internal/pkg/plugin/manager.go](./internal/pkg/plugin/manager.go)

+ mechanism for bubbling up events such as new connection/handler/listener
+ plugin manager is the only implementation
+ threaded from plugin manager to relevant children
+ used by
  + Listener
  + Handler
  + Resolver
  + Proxy

### Proxy 
secretless.Proxy

[./internal/app/secretless/proxy.go](./internal/app/secretless/proxy.go)

+ takes
  + EventNotifier
  + [Resolver](#resolver)
  + RunListenerFunc - ListenerFactory
    + takes 
      + listener id
      + ListenerOptions
    + returns Listener interface pointer
  + RunHandlerFunc - HandlerFactory, passed to listener
    + NOTE: HandlerFactories run Handlers before returning them.
    + takes 
      + handler id
      + HandlerOptions
    + returns Listener interface pointer
+ manages Listener lifecycles
  + in #Listen, for each listener 
    + creates net.Listener based on config.Config
    + creates Listener interface pointer using RunListenerFunc

### Listener 
interface: [./pkg/secretless/plugin/v1/listener.go](./pkg/secretless/plugin/v1/listener.go)

implementations: [./internal/app/secretless/listeners/](./internal/app/secretless/listeners/)

+ can be built on top of BaseListener in [./pkg/secretless/plugin/v1/listener.go](./pkg/secretless/plugin/v1/listener.go)
+ accepts client connections and passes them to relevant handler 
  + in most cases one handler that matches the protocol, http is the exception
+ manages Handler lifecycles
  + creates Handler interface pointer using RunHandlerFunc

### Handler 
interface: [./pkg/secretless/plugin/v1/handler.go](./pkg/secretless/plugin/v1/handler.go)

implementations: [./internal/app/secretless/handlers/](./internal/app/secretless/handlers/)

+ can be built on top of built on top of BaseHandler in [./pkg/secretless/plugin/v1/handler.go](./pkg/secretless/plugin/v1/handler.go)
+ receives client connection and connects to a backend and streams connection. 
+ lifecycle
  + startup
  + configure and connect to backend
  + stream
  + shutdown

### Resolver
interface: [./pkg/secretless/plugin/v1/resolver.go](./pkg/secretless/plugin/v1/resolver.go)

implementation: [./internal/pkg/plugin/resolver.go](./internal/pkg/plugin/resolver.go)

+ manages Providers and provides convenient interface to obtain multiple values from multiple secret vault backends
+ takes
  + ProviderFactories
  + EventNotifier

