// Package signal is a wrapper over the os/signal package that allows multiple
// handlers to subscribe to notifications of "exitListener" signals. Subscribers are
// guaranteed to be notified in the order in which subscriptions were made, so
// that the last subscriber will be last to be notified.
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

// ExitListener reifies the idea of "an exitListener signal" so that such an exitListener is smart: It
// allows you to add handlers that will be called when it occurs (AddHandler),
// and to block until it does occur (Wait).
type ExitListener interface {
	AddHandler(Handler)
	Wait()
}

type exitListener struct {
	handlers          []Handler
	exitSignalChannel chan os.Signal
	doneChannel       chan struct{}
}

// AddHandler adds a new subscriber to be notified when an exitListener signal is
// received. Subscribers are guaranteed to be notified in the same order the
// are added.
func (p *exitListener) AddHandler(exitHandler Handler) {
	p.handlers = append(p.handlers, exitHandler)
}

// Wait does two things: 1. It kicks off the "listening" process, so that
// Handlers will be notified of an exitListener. 2.  Blocks until an exitListener occurs.
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

// NewExitListener creates a new instance of ExitListener.  Clients are responsible for adding
// handlers and calling Wait() to kick it off.
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
