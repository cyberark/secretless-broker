/*
Package v1 plugins are ways to add extended functionality to the Secretless Broker.

Note: Given the speed of development, it is possible this documentation may become
outdated. Please use this document as a reference, but rely on the GitHub source
code as the source of truth.

Supported Plugin Types

The following types of plugins are currently supported:
  - Listeners
  - Handlers
  - Providers
  - Configuration managers
  - Connection managers

There is additionally an EventNotifier class that is currently  used to propagate
events from listeners and handlers to the plugin manager. Since it is expected
that this class may be removed as we move more of the abstract functionality to
the plugin manager itself, the EventNotifier class is only minimally covered in
this documentation.

Loading Plugins on Secretless Broker Start

When you start Secretless Broker, it will by default look in /usr/local/lib/secretless
for any plugin shared library files to load. You can specify an alternate directory
at startup by using the plugin directory flag:

  ./secretless-broker -p /etc/lib/secretless

Optionally (and highly recommended) is also providing a checksum of plugins to the broker
so that verification can be done of expected vs actual plugins and ensure that plugins have
not been modified. By doing the checksum validation any new plugins that are not expected
or plugins with modified content will prevent the broker from starting, ensuring that no
malicious (or corrupted) libraries will be loaded.

Plugin checksum verification in general should eliminate:
  - Plugin library code injections
  - Drive-by plugin content modifications
  - Addition of malicious plugins
  - Plugin corruption
  - Modification of plugins when they are mounted from other locations

You can provide the checksums in the standard "sha256sum" format that is available in most
distributions:

   sha256sum /path/to/plugins/dir/* > PLUGINS_SHA256SUM.txt
  ./secretless-broker -p /etc/lib/secretless -s PLUGINS_SHA256SUM.txt

When Secretless Broker starts, all plugins are currently loaded in the following manner:

1. The plugin directory (by default set to /usr/local/lib/secretless) is checked
for any shared library (*.so) files. Sub-directory traversal is not supported
at this time.

2. Each plugin shared library is loaded and validated to ensure it contains
all required variables and methods. In the "Plugin Minimum Requirements"
section below we discuss these requirements in greater detail. To briefly summarize,
Secretless Broker expects any plugin to include definitions for the following
variables and methods:
  - var PluginAPIVersion
  - var PluginInfo
  - func GetHandlers
  - func GetListeners
  - func GetProviders
  - func GetConfigurationManagers
  - func GetConnectionManagers

3. For each plugin, every component factory is enumerated:
  - Handler plugins are added to handler factory map
  - Listener plugins are added to listener factory map
  - Provider plugins are added to provider factory map
  - Connection manager plugins are added to connection manager factory map
  - Configuration manager plugins are added to configuration manager factory map

4. Connection manager plugins are instantiated

5. The chosen configuration manager plugin is instantiated

6. The program waits for a valid configuration to be provided

7. After the configuration is provided and loaded, providers and listeners/handlers
are instantiated as needed

Plugin Minimum Requirements

In this section, we will go over the minimum requirements for the variables and
methods that every custom plugin must include in order to be properly loaded
into the Secretless Broker.

  // PluginAPIVersion is the target API version of the Secretless Broker and must
  // match the supported version defined in
  // internal/plugin/manager.go:_IsSupportedPluginAPIVersion
  string PluginAPIVersion

  // PluginInfo is a map that has information about the plugin that the daemon
  // might use for logging, prioritization, and masking.
  // While extraneous keys in the map are ignored, the map must contain the
  // following keys:
  // - "version": indicates the version of the plugin
  // - "id": a computer-friendly plugin ID. Naming should be constrained to
  //         short, spaceless ASCII lowercase alphanumeric set with a limited
  //         set of special characters (`-`, `_`, and `/`).
  // - "name": a user-friendly plugin name. This name will be used in most
  //           user-facing messages about the plugin and should be constrained
  //           in length to <30 chars.
  // - "description": a plugin description; should be less than 30 characters.
  map[string]string PluginInfo

  // GetListeners returns a map of listener IDs to their factory methods
  // The factory methods accept v1.ListenerOptions when invoked and return a
  // new v1.Listener
  func GetListeners() map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener

  // GetHandlers returns a map of handler IDs to their factory methods
  // The factory methods accept v1.HandlerOptions when invoked and return a
  // new v1.Handler
  func GetHandlers() map[string]func(plugin_v1.HandlerOptions) plugin_v1.Handler

  // GetProviders returns a map of provider IDs to their factory methods
  // The factory methods accept v1.ProviderOptions when invoked and return a
  // new v1.Provider (and/or an error)
  func GetProviders() map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error)

  // GetConnectionManagers returns a map of connection manager IDs to their
  // factory methods
  // The factory methods return a new v1.ConnectionManager when invoked
  // Note: it is expected that the factory methods will also eventually have
  //       v1.ConnectionManagerOptions as an argument when invoked, for
  //       eventual consistency with the other factory maps
  func GetConnectionManagers() map[string]func() plugin_v1.ConnectionManager

  // GetConfigurationManagers returns a map of configuration manager IDs to their
  // factory methods
  // The factory methods accept v1.ConfigurationManagerOptions when invoked and
  // return a new v1.ConfigurationManager
  func GetConfigurationManagers() map[string]func(plugin_v1.ConfigurationManagerOptions) plugin_v1.ConfigurationManager

Example plugin

The following shows a sample plugin that conforms to the expected API. The full
sample plugin is available in GitHub at https://github.com/cyberark/secretless-broker/tree/master/test/plugin/example.

  package main

  import (
    plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
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

*/
package v1
