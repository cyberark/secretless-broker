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
	Wait()
}

type exitListener struct {
	handlers          []Handler
	exitSignalChannel chan os.Signal
	doneChannel       chan struct{}
}

// AddHandler adds a new handler that will be invoked when an exit signal is
// received. Handlers are invoked in the order they were added.
func (p *exitListener) AddHandler(exitHandler Handler) {
	p.handlers = append(p.handlers, exitHandler)
}

// Wait does two things: 1. It kicks off the "listening" process, so that
// Handlers will be notified of an exit event. 2.  Blocks until an
// exit signal is received.
func (p *exitListener) Wait() {
	go func() {
		<-p.exitSignalChannel

		for _, h := range p.handlers {
			h()
		}

		p.doneChannel <- struct{}{}
	}()

	<- p.doneChannel
}

// NewExitListener creates a new instance of ExitListener.  Clients are
// responsible for adding handlers and calling Wait() to kick it off.
func NewExitListener() ExitListener {
	doneChannel := make(chan struct{})
	exitSignalChannel := make(chan os.Signal)
	signal.Notify(exitSignalChannel, exitSignals...)

	return &exitListener{
		handlers:          []Handler{},
		exitSignalChannel: exitSignalChannel,
		doneChannel:       doneChannel,
	}
}
