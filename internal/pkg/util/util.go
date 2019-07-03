package util

import (
	"log"
	"net"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
)

// Accept listens for new connections from Listener `l` and notifies plugins
// of new connections
func Accept(l plugin_v1.Listener) (net.Conn, error) {
	conn, err := l.GetListener().Accept()
	if conn != nil && err == nil {
		l.GetNotifier().NewConnection(l, conn)
	}
	return conn, err
}

// OptionalDebug returns a function that will noop when debugEnabled=false, and
// will log the given msg using `log.Print` or `log.Printf` (depending on the
// number arguments given) debugEnabled=true
func OptionalDebug(debugEnabled bool) func(string, ...interface{}) {
	if !debugEnabled {
		// return a noop function
		return func(msg string, args ...interface{}) {
			return
		}
	}
	// return the real debug function
	return func(msg string, args ...interface{}) {
		if args == nil {
			log.Print(msg)
		}
		log.Printf(msg, args...)
	}
}
