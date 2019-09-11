/*
Package connector extends the base functionality of Secretless broker

General Overview

Secretless has built-in support for Postgres, MySQL, basic http authentication,
and many other target services. But what if you want to use Secretless with a
target service it doesn't support out of the box?

Service connector plugins let you extend Secretless to support any target
service. If you know Go and understand your target service's authentication
protocol, you can write a connector for it. It's incredibly easy, and this
guide walks you through it.

The Secretless team is continually adding support for new databases and
services, but we love and encourage outside contributions as well. If you
write a connector plugin you'd like to share with the community, please send us
a PR!

Technical Overview

Secretless uses Go Plugins. If you've never used Go plugins before, a good
introduction to them is at https://tinyurl.com/yaprrcsm. The short version
is that you'll write normal Go functions, but compile them with different flags.
This will produce a shared object library file (with a ".so" extension) instead
of a normal executable.

Secretless Plugins

Technically, a Secretless service connector plugin is just a Go shared library
file that implements either the "tcp.Plugin" interface or the "http.Plugin"
interface:

From github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp:

	package tcp

	type Plugin interface {
	  PluginInfo() map[string]string
	  NewConnector(connector.Resources) Connector
	}

type Connector func(net.Conn, plugin.CredentialValues) (net.Conn, error)

From github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http:

	package http

	type Plugin interface {
	  PluginInfo() map[string]string
	  NewConnector(connector.Resources) Connector
	}

	type Connector func(*http.Request, plugin.CredentialValues) error

We'll get into the details below, but at a high-level, your connector itself --
that is, the thing returned by NewConnector -- is simply a function that
knows how to transform an unauthenticated connection or request into an
authenticated one.

To get that job done, Secretless provides you with connector.Resources
(detailed below) as well as the current credential values --  the secrets
you'll need to authenticate. Your plugin users specify the location of those
secrets in secretless.yml, as described at
https://docs.secretless.io/Latest/en/Content/References/connectors/overview.htm#ConfigureSecretlesstolistenfornewconnections.
At runtime, Secretless fetches the values of those secrets and passes
the into your Connector function.

Deploying Secretless Plugins

After compiling your plugin code into a ".so" file, you'll place the ".so" file
in the "/usr/local/lib/secretless" directory in the container where Secretless
will run. That's all you have to do. The ".so" files are self-contained, with
any dependencies you've imported baked in.

Note: ".so" plugin files must be placed directly in
"/usr/local/lib/secretless". Sub-directories of that folder aren't searched.

Securing Plugin Deployments

Plugins can be secured by a checksum file to prevent injection attacks. We
recommend all production deployments use this feature. Find out more
at https://github.com/cyberark/secretless-broker/blob/master/internal/plugin/v1/doc.go.

Technical Details

This section details the interfaces and types you'll need to implement or use
when authoring a plugin.

PluginInfo()

This is one of the two methods required by both the "tcp.Plugin" and
"http.Plugin" interfaces. It returns basic information about your plugin. It's
signature is:

	func PluginInfo() map[string]string

The returned map must have the following keys:

	- version: The version of the plugin itself. This allows plugin authors to
	  version the plugins they write.
	- pluginAPIVersion: The version of the Secretless plugin API your plugin is
	  written for. This allows the Secretless plugin API to change over time
	  without breaking plugins.
	- id: A short, clear, unique name, for use in logs and the "secretless.yml"
	  config file. Allowed characters are: lowercase letters, "_", ":", "-", and
	  "~".
	- description: A short summary of the plugin, not to exceed 100 characters.
	  This may be used in the future by the secretless cmd line tool to list
	  available plugins.

NewConnector

In both the "tcp.Plugin" and "http.Plugin" interfaces, NewConnector returns a
Connector, which in turn is a function that performs the actual
authentication. The function signature is slightly different depending on
which interface you're implementing.

When Secretless runs, it will call NewConnector once, and then hold onto the
returned Connector. That Connector (remember: it's just a function) will
then be called each time a new client connection requires authentication.

Both NewConnector methods take only one argument: "connector.Resources" which
is described below.

The real work is done by the Connector functions they return...

TCP Connector

This is the function returned by tcp.Plugin's NewConnector(), and it's where your
TCP authentication logic lives. It's signature is:

	func(clientConn net.Conn, credVals plugin.CredentialValues)
	    (authdTargetServiceConn net.Conn, err error)

That is, it's passed the client's net.Conn and the current
CredentialValues, and returns an authenticated net.Conn to the target
service. The authentication stage is complete after Connector is called.

Secretless will now have both the client connection and an authenticated
connection to the target service. The relationship between the client
connection, Secretless, and the authenticated target service connection looks
like this:

	clientConn <--> Secretless <--> authdTargetServiceConn

At this point, Secretless becomes an invisible proxy, streaming bytes back and
forth between client and target service, as if they were directly
connected.

HTTP Connector

This is the function returned by http.Plugin's NewConnector(), and it's
where the http authentication logic lives. It's signature is:

	func(*http.Request, credVals plugin.CredentialValues) error

Here we are passed a pointer to an http.Request and CredentialValues, and
are expected to alter that request so that it contains the necessary
authentication information. Typically, this means adding the appropriate
headers to a request -- for example, an Authorization header containing a
Token, or a header containing an API key.

Since HTTP is a stateless protocol, Secretless will call this function every
time a client sends an http request to the target server, so that every request
will be authenticated.

connector.Resources

Everything that your Connector needs from the Secretless framework is exposed
through the connector.Resources interface, which is passed to your plugin's
constructor. You will need to retain a reference to connector.Resources, via
a closure, inside the Connector function returned by your constructor.
Here's the connector.Resources interface:

package connector

	type Resources interface {
	  Config()          []byte
	  Logger()          secretless.Logger
	}

Logger() provides a basic logger you can use for debugging and informational
logging. Config() provide you resources specified in your secretless.yml
file.

Let's break down each method:

	- Config() - Some connectors require additional, connector-specific
	  configuration. Anything specified in your connector's "config" section will
	  be passed back via this method as a raw []byte. Your code is responsible
	  for casting those bytes back into a meaningful "struct" that your code can
	  work with.
	- Logger() - Returns an object similar to the standard library's
	  log.Logger. This lets you log events to stdout and stderr. It respects
	  Secretless's command line "debug" flag, so that calling Debugf or Infof
	  will do nothing unless you started Secretless in debug mode. See below for
	  details.

secretless.Logger Interface

Your code should never use the "fmt" or "log" packages, or write directly to
stdout or stderr. Instead, call Logger() on your connector.Resources to get
a secretless.Logger with the following interface:

	type Logger interface {
	  Debugf(format string, v ...interface{})
	  Debug(msg string)
	  Debugln(msg string)

	  Infof(format string, v ...interface{})
	  Info(msg string)
	  Infoln(msg string)

	  Warnf(format string, v ...interface{})
	  Warn(msg string)
	  Warnln(msg string)

	  Errorf(format string, v ...interface{})
	  Error(msg string)
	  Errorln(msg string)

	  Fatalf(format string, v ...interface{})
	  Fatal(msg string)
	  Fatalln(msg string)
	}

The 3 Debug methods and the 3 Info methods do nothing unless you started
Secretless with the Debug command line flag set to true. If you did start
Secretless in debug mode, they write to stdout.

The Warn, Error, and Fatal methods all write to stderr, regardless of
the current Debug mode.

All the Fatal methods call os.Exit(1) after printing their message.

The Logger automatically prepends the name of the currently running service
to all messages. That is, the service name specified in your secretless.yml.
For example, if your secretless.yml looks like:

	version: "v2"
	services:
	  sample-service:
	    protocol: pg
	    listenOn: unix:///sock/.s.PGSQL.5432
	    ...

then the Logger will prepend sample-service to all messages.

Examples

	TBD

Testing

	TBD

*/
package connector
