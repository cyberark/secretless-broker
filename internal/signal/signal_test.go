package signal_test

import (
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/signal"
)

func TestSignal(t *testing.T) {

	t.Run("Handlers handle expected signals", func(t *testing.T) {
		// SIGUSR1 won't interfere with the process, so it's a good signal to
		// test with.
		expectedSignal := syscall.SIGUSR1
		exitListener := signal.NewExitListenerWithOptions(expectedSignal)

		// Create a channel for the handlers to communicate with.  Make the
		// channel size larger than expected result size, so if we get too many
		// results, we'll know.
		handlerResultsCh := make(chan string, 99)

		// Create and add the handlers
		handler1 := func() {
			handlerResultsCh <- "handler1"
		}
		handler2 := func() {
			handlerResultsCh <- "handler2"
		}
		exitListener.AddHandler(handler1)
		exitListener.AddHandler(handler2)

		// doneCh will be notified after exitListener has handled signals
		doneCh := exitListener.Listen()

		// Fire the signal
		syscall.Kill(syscall.Getpid(), expectedSignal)

		// Add timeout so test fails cleanly if expected signal isn't handled.
		select {
		case <-doneCh:
		case <-time.After(200 * time.Millisecond):
			assert.FailNow(t, "exitListener failed to handle expected signal")
		}

		// Get the results
		var results []string
		results = append(results, <-handlerResultsCh)
		results = append(results, <-handlerResultsCh)

		// Confirm the expected results are correct
		assert.EqualValues(t, []string{"handler1", "handler2"}, results)

		// Confirm the channel is empty (no unexpected results)
		assert.Equal(t, 0, len(handlerResultsCh))
	})

	t.Run("Handlers don't handle unexpected signals", func(t *testing.T) {
		// SIGUSR1 won't interfere with the process, so it's a good signal to
		// test with.
		expectedSignal := syscall.SIGUSR1
		unexpectedSignal := syscall.SIGUSR2
		exitListener := signal.NewExitListenerWithOptions(expectedSignal)

		// doneCh will be notified after exitListener has handled signals
		doneCh := exitListener.Listen()

		// Fire the signal
		syscall.Kill(syscall.Getpid(), unexpectedSignal)

		// If incorrectly handling doesn't occur in 200ms, we're safe.
		select {
		case <-doneCh:
			assert.FailNow(t, "exitListener incorrectly handled unexpected signal")
		case <-time.After(200 * time.Millisecond):
		}
	})
}
