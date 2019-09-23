// Package signal is a wrapper over the os/signal package that allows multiple
// handlers to subscribe to notifications of "exit" signals. Subscribers are
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

// Exit reifies the idea of "an exit signal" so that such an exit is smart: It
// allows you to add handlers that will be called when it occurs (AddHandler),
// and to block until it does occur (Await).
type Exit interface {
	AddHandler(Handler)
	Await()
}

type exit struct {
	handlers          []Handler
	exitSignalChannel chan os.Signal
	doneChannel       chan struct{}
}

// AddHandler adds a new subscriber to be notified when an exit signal is
// received. Subscribers are guaranteed to be notified in the same order the
// are added.
func (p *exit) AddHandler(exitHandler Handler) {
	p.handlers = append(p.handlers, exitHandler)
}

// Await does two things: 1. It kicks off the "listening" process, so that
// Handlers will be notified of an exit. 2.  Blocks until an exit occurs.
func (p *exit) Await() {
	go func() {
		<-p.exitSignalChannel

		for _, sub := range p.handlers {
			sub()
		}

		p.doneChannel <- struct{}{}
	}()

	<- p.doneChannel
}

// NewExit creates a new instance of Exit.  Clients are responsible for adding
// handlers and calling Await() to kick it off.
func NewExit() Exit {
	doneChannel := make(chan struct{})
	exitSignalChannel := make(chan os.Signal) //TODO: should this be 0?
	signal.Notify(exitSignalChannel, exitSignals...)

	return &exit{
		handlers:          []Handler{},
		exitSignalChannel: exitSignalChannel,
		doneChannel:       doneChannel,
	}
}
