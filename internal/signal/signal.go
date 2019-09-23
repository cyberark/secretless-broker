// Package signal is a wrapper over the os/signal package that allows multiple
// subscribers to subscribe to notifications of "exit" signals. Subscribers are
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

// Subscriber represents anything than wants to be notified of exit signals.
// It's simply a function that takes no arguments.  Actual subscribers will
// typically use this function as a closure to wrap up relevant logic at
// the call site.
type Subscriber func()

// Publisher defines the interface of a publisher.
type Publisher interface {
	Subscribe(Subscriber)
	Start()
	Stop()
}

type exitSignalPublisher struct {
	subscribers   []Subscriber
	signalChannel chan os.Signal
	doneChannel   chan struct{}
}

// Subscribe adds a new subscriber to be notified when an exit signal is
// received. Subscribers are guaranteed to be notified in the same order the
// are added.
func (p *exitSignalPublisher) Subscribe(subscriber Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

// Start must be called to kick off the publication process.
func (p *exitSignalPublisher) Start() {
	go func() {
		for {
			select {
			case <-p.signalChannel:
				for _, sub := range p.subscribers {
					sub()
				}
			case <-p.doneChannel:
				return
			}
		}
	}()
}

// Stop makes the publisher stop listening for and publishing signal events.
func (p *exitSignalPublisher) Stop() {
	p.doneChannel <- struct{}{}
}

// NewExitSignalPublisher creates a new instance of publisher.  The publisher
// will not start actually publishing exit signal events until Start is called.
func NewExitSignalPublisher() Publisher {
	doneChannel := make(chan struct{})
	signalChannel := make(chan os.Signal, 1) //TODO: should this be 0?
	signal.Notify(signalChannel, exitSignals...)

	return &exitSignalPublisher{
		subscribers:   []Subscriber{},
		signalChannel: signalChannel,
		doneChannel: doneChannel,
	}
}
