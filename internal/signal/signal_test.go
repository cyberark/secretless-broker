package signal_test

import (
	"os"
	ossignal "os/signal"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/signal"
)

func TestSignal(t *testing.T) {
	t.Run("Signals listeners are notified", func(t *testing.T) {
		testSignals := []os.Signal{
			syscall.SIGUSR1,
			syscall.SIGUSR2,
		}
		exitListener := signal.NewExitListenerWithOptions(testSignals...)

		notifications := []string{}
		handler1 := func() {
			notifications = append(notifications, "handler1")
		}
		handler2 := func() {
			notifications = append(notifications, "handler2")
		}
		handler3 := func() {
			notifications = append(notifications, "handler3")
		}

		exitListener.AddHandler(handler1)
		exitListener.AddHandler(handler2)
		exitListener.AddHandler(handler3)

		// Sanity check
		assert.Equal(t, 0, len(notifications))

		go func() {
			syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
		}()

		exitListener.Wait()

		assert.Equal(t, []string{"handler1", "handler2", "handler3"}, notifications)
	})

	t.Run("Signals listeners are only notified on expected signals", func(t *testing.T) {
		testSignals := []os.Signal{
			syscall.SIGUSR1,
		}
		exitListener := signal.NewExitListenerWithOptions(testSignals...)

		notifications := []string{}
		handler1 := func() {
			notifications = append(notifications, "handler1")
		}
		handler2 := func() {
			notifications = append(notifications, "handler2")
		}

		exitListener.AddHandler(handler1)
		exitListener.AddHandler(handler2)

		// Start the ExitListener and create a channel it will notify when it's
		// done listening.
		waitIsFinished := make(chan struct{})
		go func() {
			exitListener.Wait()
			waitIsFinished <- struct{}{}
		}()

		// Sanity check
		assert.Equal(t, 0, len(notifications))

		// It responds to signals it's listening for
		syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

		// Wait for the ExitListener to respond
		<-waitIsFinished

		// Expected notification was added
		assert.Equal(t, 2, len(notifications))

		// Now the 2nd case, that it ignores signals it's not listening for
		// NOTE: This will be moved to a separate test

		// Reset the notifications
		notifications = []string{}

		// First we setup a channel so we'll know when the signal has been
		// processed. signal package notifies in order of subscription, so if
		// this is called, we know ExitListener will have already been called.
		// Nit: It's _technically_ possible ExitListener was called but isn't
		// processing yet, but this is very unlikely.
		ignoredSignal := syscall.SIGUSR2
		signalProcessedCh := make(chan os.Signal, 1)
		ossignal.Notify(signalProcessedCh, ignoredSignal)

		go func() {
			exitListener.Wait()
			waitIsFinished <- struct{}{}
		}()

		// Send the ignored signal
		syscall.Kill(syscall.Getpid(), ignoredSignal)

		// Wait till we know it's been processed
		<-signalProcessedCh

		// Now assert that ExitListener did NOT handle it
		assert.Equal(t, 0, len(notifications))
	})
}
