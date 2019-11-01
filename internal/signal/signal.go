// Package signal is a wrapper over the os/signal package that allows multiple
// handlers to respond to an exit signal, and blocks until that exit signal is
// received.
package signal

import (
	"os"
	"os/signal"
	"syscall"
)

var exitSignals = []os.Signal{
	syscall.SIGABRT,
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGTERM,
}

// Handler is a simply a function intended to be called in response to a signal.
type Handler func()

// ExitListener listens for exit signals, and responds by invoking any handlers
// that have been added.  It start listening until Wait() in invoked, at which
// point it will block until it receives any exit signal, call the handlers in
// the order they were added, and then stop listening.
type ExitListener interface {
	AddHandler(Handler)
	Listen() chan struct{}
}

type exitListener struct {
	handlers          []Handler
	exitSignalChannel chan os.Signal
}

// AddHandler adds a new handler that will be invoked when an exit signal is
// received. Handlers are invoked in the order they were added.
func (p *exitListener) AddHandler(exitHandler Handler) {
	p.handlers = append(p.handlers, exitHandler)
}

// Listen does two things: 1. It kicks off the "listening" process, so that
// Handlers will be notified of an exit event. 2.  Returns a channel that
// will be notified only after those events have been handled.
//
// NOTE: Listen should only be called once.  Create a new listener if you need
// to call it again.
func (p *exitListener) Listen() chan struct{} {
	doneChannel := make(chan struct{})
	go func() {
		<-p.exitSignalChannel

		for _, h := range p.handlers {
			h()
		}

		doneChannel <- struct{}{}
	}()
	return doneChannel
}

// NewExitListener creates a new instance of ExitListener.  Clients are
// responsible for adding handlers and calling Wait() to kick it off.
func NewExitListener() ExitListener {
	return NewExitListenerWithOptions(exitSignals...)
}

// NewExitListenerWithOptions creates a new instance of ExitListener with configurable
// options. Clients are responsible for adding handlers to the listener which can then
// be waited on by `Wait()`ing.
func NewExitListenerWithOptions(signals ...os.Signal) ExitListener {
	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, signals...)

	return &exitListener{
		handlers:          []Handler{},
		exitSignalChannel: exitSignalChannel,
	}
}
