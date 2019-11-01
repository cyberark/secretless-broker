package signal_test

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/signal"
)

func TestSignal(t *testing.T) {

	t.Run("Signals listeners are notified on expected signals", func(t *testing.T) {
		// SIGUSR1 won't interfere with the process, so it's a good signal to
		// test with.
		expectedSignal := syscall.SIGUSR1
		exitListener := signal.NewExitListenerWithOptions(expectedSignal)

		// Create a channel for the handlers to communicate with.
		handlerResultsCh := make(chan string, 2)

		// Create and add the handlers
		handler1 := func() {
			handlerResultsCh <- "handler1"
		}
		handler2 := func() {
			handlerResultsCh <- "handler2"
		}
		exitListener.AddHandler(handler1)
		exitListener.AddHandler(handler2)

		// Start the ExitListener and create a channel it will notify when it's
		// done listening.
		waitHasReturnedCh := make(chan struct{})
		go func() {
			exitListener.Wait()
			waitHasReturnedCh <- struct{}{}
		}()

		// Ensure that Wait() has started
		<-exitListener.IsWaitingCh()

		// Fire the expected signal
		syscall.Kill(syscall.Getpid(), expectedSignal)

		// Wait for the ExitListener to respond
		<-waitHasReturnedCh

		// Get the results
		var results []string
		results = append(results, <-handlerResultsCh)
		results = append(results, <-handlerResultsCh)

		// Confirm they're correct
		assert.EqualValues(t, []string{"handler1", "handler2"}, results)
	})

	t.Run("Handlers handle expected signals and ignore others", func(t *testing.T) {
		// SIGUSR1 won't interfere with the process, so it's a good signal to
		// test with.
		expectedSignal := syscall.SIGUSR1
		unexpectedSignal := syscall.SIGUSR2
		exitListener := signal.NewExitListenerWithOptions(expectedSignal)

		// Create a channel for the handlers to communicate with.
		//
		// NOTE: We give a buffer large enough to hold more results than we're
		// expecting if the implementation is correct (2 results).  Because if
		// the unexpected signal were handled (the bug we're testing for), we
		// need space to catch both its results and the expected results.
		handlerResultsCh := make(chan string, 4)

		// Create and add the handlers
		handler1 := func() {
			handlerResultsCh <- "handler1"
		}
		handler2 := func() {
			handlerResultsCh <- "handler2"
		}
		exitListener.AddHandler(handler1)
		exitListener.AddHandler(handler2)

		// Start the ExitListener and create a channel it will notify when it's
		// done listening.
		waitHasReturnedCh := make(chan struct{})
		go func() {
			exitListener.Wait()
			waitHasReturnedCh <- struct{}{}
		}()

		// Ensure that Wait() has started
		<-exitListener.IsWaitingCh()

		// Fire the unexpected signal first, then the expected one.  If there
		// expected one is handled, we know for sure the unexpected would have
		// been handled, _if_ such a bug existed.  That is, we can use it as
		// a synchronization mechanism.
		syscall.Kill(syscall.Getpid(), unexpectedSignal)
		syscall.Kill(syscall.Getpid(), expectedSignal)

		// Wait for the ExitListener to respond
		<-waitHasReturnedCh

		// Get the results
		var results []string
		results = append(results, <-handlerResultsCh)
		results = append(results, <-handlerResultsCh)

		// Confirm the expected results are correct
		assert.EqualValues(t, []string{"handler1", "handler2"}, results)

		// Confirm the channel is empty (no unexpected results)
		assert.Equal(t, 0, len(handlerResultsCh))
	})
}
