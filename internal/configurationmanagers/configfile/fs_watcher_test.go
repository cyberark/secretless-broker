package configfile

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAttachWatcher(t *testing.T) {
	t.Run("AttachWatcher", func(t *testing.T) {
		// Mock the log.Fatal function to intercept fatal errors for inspection
		var watcherErr error
		logFatal = func(v ...interface{}) {
			watcherErr = v[0].(error)
		}
		// Create a listener that will track the number of times the watcher detects a file change
		changeCount := 0
		// Allow a 10ms delay for the watcher to detect the file change
		delay := time.Millisecond * 10
		onChange := func() {
			changeCount++
		}

		checkChangeCount := func(t *testing.T, expectedValue int) {
			// Wait the specified amount, the assert that the change count is as expected
			time.Sleep(delay)
			assert.Equal(t, expectedValue, changeCount)
		}

		// Create a temp file to watch
		tempDir, _ := ioutil.TempDir("", "configfile_watcher")
		defer os.RemoveAll(tempDir)
		file, err := os.CreateTemp(tempDir, "configfile")
		assert.NoError(t, err)

		AttachWatcher(file.Name(), onChange)
		checkChangeCount(t, 0)
		assert.NoError(t, watcherErr)

		// Write to the file and check that the listener is called
		_, err = file.WriteString("test")
		assert.NoError(t, err)
		checkChangeCount(t, 1)
		assert.NoError(t, watcherErr)

		_, err = file.WriteString("test again")
		file.Close()
		assert.NoError(t, err)
		checkChangeCount(t, 2)
		assert.NoError(t, watcherErr)

		// Delete the file and check that the listener is called and an error occurs
		err = os.Remove(file.Name())
		assert.NoError(t, err)
		// Wait longer because the watcher tries to reattach to deleted files in case they're recreated
		time.Sleep(1 * time.Second)
		checkChangeCount(t, 3)
		assert.Error(t, watcherErr)
	})
}
