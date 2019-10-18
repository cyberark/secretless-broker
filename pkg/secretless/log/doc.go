/*
Package log provides an interface similar to the standard library "log.Logger".  It allows you to
log events to stdout and stderr.  It respects the Secretless command line `debug` flag,
so that calling Debugf or Infof does nothing unless you started Secretless in
debug mode.

Secretless plugins should never use the "fmt" or "log" packages or write
directly to stdout or stderr.  Instead, get a secretless.Logger as defined below
by calling the Logger() method on the connector.Resources input to your
NewConnector function.

The three Debug methods and the three Info methods do nothing unless you started
Secretless with the Debug command line flag set to "true".  If you did start
Secretless in debug mode, the methods write to stdout.

The Warn, Error, and Fatal methods all write to stderr, regardless of
the current Debug mode.

All of the Fatal methods call `os.Exit(1)` after printing their message.

The `Logger` automatically prepends the service name specified in your `secretless.yml`
for the currently-running service to all messages. For example, if your
`secretless.yml` looks like:


  version: "v2"
  services:
    sample-service:
      protocol: pg
      listenOn: unix:///sock/.s.PGSQL.5432
      ...


then the Logger prepends "sample-service" to all messages.
*/
package log
