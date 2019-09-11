// Package signal is a thin wrapper wrapper over the os/signal package that
// allows any type with a Stop() method to be conveniently stopped when any
// of the standard os "halt" signals are raised.
//
// The package only exposes a single method StopOnExitSignal(Stopper).  Just
// pass any Stopper to that function, and its Stop() method will be called
// when a halt signal is raised.
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

type Stopper interface {
	Stop()
}

// newHaltSignalChan returns a new channel containing any "halt"-like signal.
// See exitSignals for a definition of a "halt"-like signal.
func newHaltSignalChan() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, exitSignals...)
	return signalChannel
}

// StopOnExitSignal will make any "Stopper" automatically Stop() when OS kill
// or kill like signals are received.
// TODO: Possible synchronization to ensure all go routines complete.
// TODO: Current usage in main relies on the fact that these will be triggered
//   in the order they're setup -- an assumption that relies on impl details of
//   the signal package.  Not ideal.
func StopOnExitSignal(s Stopper) {
	killSignals := newHaltSignalChan()

	go func() {
		<-killSignals
		s.Stop()
	}()
}
