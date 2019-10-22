# Secretless Plugins

Secretless plugins allow you to extend the functionality of Secretless beyond the
currently supported built-in plugins.

## Currently Supported Plugin Types

Secretless supports plugins for the following internal components:

  - [Service Connectors](connector/README.md)

## External Plugin Basics

Secretless uses [Go plugins](https://golang.org/pkg/plugin/).

If you've never used Go plugins before, a good introduction to them is
[here](https://medium.com/learning-the-go-programming-language/writing-modular-go-programs-with-plugins-ec46381ee1a9).
Essentially, to write a Secretless plugin you'll write normal Go functions but compile them
using `-buildmode=plugin`. This produces a shared object library file (with a
`.so` extension) instead of a normal executable.

Technically, a Secretless plugin is a Go shared library
file that implements some predefined functions. For more information on what you
need to implement to build a plugin, please see the README for the specific
[plugin type](#currently-supported-plugin-types) you are building.

## Plugin Metadata

Regardless of plugin type, each plugin must supply Secretless with some essential metadata.
To do this, each plugin must implement the `PluginInfo` function. This top level function is
always required and it returns basic information about your plugin. Its signature is:

```go
func PluginInfo() map[string]string
```

The returned map must have the following keys:

- `version`: The version of the plugin itself.  This allows plugin authors to
  version the plugins they write.
- `pluginAPIVersion`: The version of the Secretless plugin API that your plugin is
  written for.  This allows the Secretless plugin API to change over time
  without breaking plugins.
  <!-- TODO: how can a plugin dev find the appropriate version to use? -->
- `type`: This must be a supported plugin type. Currently, it must be
  either the string `"connector.tcp"` or the string `"connector.http"`.
- `id`: A short, clear, unique name for use in logs and the `secretless.yml`
  config file.  Allowed characters are: lowercase letters, `_`, `:`, `-`, and
  `~`.
- `description`: A short summary of the plugin, not to exceed 100 characters.
  This may be used in the future by the Secretless command line tool to list
  available plugins.

## External Plugin Basics

When running Secretless with external plugins, you can leverage some special
command-line flags when starting Secretless:

- `-p` flag: Specifies the directory in which the external plugins shared library
  (".so") files live.. Defaults to `/usr/local/lib/secretless`. Sub-directory traversal
  is not supported at this time.
- `-s` flag: Refers to a file that contains sha256sum plugin checksums for verifying the plugins.

When Secretless starts, it:
- Checks for available external plugins (eg ".so" files) in the plugin directory.
- Verifies external plugin checksums (if a checksum file was provided on start).
- Loads the external plugin. For each plugin file Secretless:
  - Opens the [Go plugin](https://golang.org/pkg/plugin/) file.
  - Parses `PluginInfo` for [plugin metadata](#plugin-metadata).
  - Verifies that the plugin type supplied in `PluginInfo` is supported.
  - Loads the plugin into the list of plugins to run.

From there, the startup process continues and external plugins are treated the same
as internal plugins.

### Building the Shared Library File
To build your plugin's shared library (`.so`) file, follow the [instructions](https://golang.org/pkg/plugin/)
for building Go plugins.

For example, to compile your plugin code into a `.so` file, run the following command:
```
go build -buildmode=plugin -o=/path/to/my-plugin.so my_plugin.go
```

Once you've done this, place the `.so` file in the `/usr/local/lib/secretless`
directory in the container where Secretless will run (or in another directory you
specify using the `-p` flag). That's all you have to do. The `.so` files are
self-contained and include any dependencies that you've imported.

*Note: `.so` plugin files must be placed directly in `/usr/local/lib/secretless`
(or the directory you specify). Sub-directories of the plugin folder are not searched.*


### Plugin Checksum Verification
Plugins can be secured by a checksum file to prevent injection attacks. We recommend
all production deployments use this feature. Find out more [here](https://github.com/cyberark/secretless-broker/blob/78552bd3065a5b11b93a6f3e20f9f5309e7d5112/internal/app/secretless/plugin/v1/doc.go#L31).
