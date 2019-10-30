package signal_test

import (
	"os"
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
		mockListener := signal.NewExitListenerWithOptions(testSignals...)

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

		mockListener.AddHandler(handler1)
		mockListener.AddHandler(handler2)
		mockListener.AddHandler(handler3)

		// Sanity check
		assert.Equal(t, 0, len(notifications))

		go func() {
			syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
		}()

		mockListener.Wait()

		assert.Equal(t, []string{"handler1", "handler2", "handler3"}, notifications)
	})

	t.Run("Signals listeners are only notified on expected signals", func(t *testing.T) {
		testSignals := []os.Signal{
			syscall.SIGUSR1,
			syscall.SIGUSR1,
		}
		mockListener := signal.NewExitListenerWithOptions(testSignals...)

		notifications := []string{}
		handler1 := func() {
			notifications = append(notifications, "handler1")
		}
		handler2 := func() {
			notifications = append(notifications, "handler2")
		}

		mockListener.AddHandler(handler1)
		mockListener.AddHandler(handler2)

		// Sanity check
		assert.Equal(t, 0, len(notifications))

		go func() {
			syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
		}()

		go func() {
			assert.Equal(t, 0, len(notifications))
			syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
		}()

		// This should block until we verify that nothing was sent
		mockListener.Wait()
	})
}
