package command

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func splitEq(s string) (string, string) {
	a := strings.SplitN(s, "=", 2)
	return a[0], a[1]
}

type envSnapshot struct {
	env []string
}

func clearEnv() *envSnapshot {
	e := os.Environ()

	for _, s := range e {
		k, _ := splitEq(s)
		os.Setenv(k, "")
	}
	return &envSnapshot{env: e}
}

func (e *envSnapshot) restoreEnv() {
	clearEnv()
	for _, s := range e.env {
		k, v := splitEq(s)
		os.Setenv(k, v)
	}
}

func assertMissingFile(f string, t *testing.T) {
	_, err := os.Stat(f)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func assertFileContents(f string, expectedValue string, t *testing.T) {
	actualContent, err := os.ReadFile(f)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, string(actualContent))
}

func TestTempFactory_Cleanup(t *testing.T) {
	t.Run("Cleanup deletes all temp files", func(t *testing.T) {
		tempFactory := NewCustomTempFactory("", "non-existent")

		f1, err := tempFactory.Push("meow")
		assert.NoError(t, err)
		f2, err := tempFactory.Push("moo")
		assert.NoError(t, err)

		assertFileContents(f1, "meow", t)
		assertFileContents(f2, "moo", t)

		tempFactory.Cleanup()

		assertMissingFile(f1, t)
		assertMissingFile(f2, t)
	})
}

func TestTempFactory_Push(t *testing.T) {
	t.Run("Push creates temp file", func(t *testing.T) {
		tempFactory := NewTempFactory("")
		defer tempFactory.Cleanup()

		f, err := tempFactory.Push("moo")
		assert.NoError(t, err)
		assertFileContents(f, "moo", t)
	})

	t.Run("Push reports errors", func(t *testing.T) {
		tempFactory := NewTempFactory("dir-not-found")
		defer tempFactory.Cleanup()

		_, err := tempFactory.Push("moo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})
}

func TestTempFactory_NewTempFactory(t *testing.T) {
	t.Run("Uses constructor arg path if provided", func(t *testing.T) {
		tempFactory := NewTempFactory("somedir")
		defer tempFactory.Cleanup()

		assert.ObjectsAreEqualValues(TempFactory{
			files: []string(nil),
			path:  "somedir",
		}, tempFactory)
	})

	t.Run("When constructor path is not provided", func(t *testing.T) {
		env := clearEnv()
		defer env.restoreEnv()

		t.Run("tries using shared memory path first", func(t *testing.T) {
			tempFactory := NewTempFactory("")

			_, err := os.Stat("/dev/shm")
			if os.IsNotExist(err) {
				return
			}

			assert.ObjectsAreEqualValues(TempFactory{
				files: []string(nil),
				path:  "/dev/shm",
			}, tempFactory)
		})

		t.Run("tries using homedir prefix if shared memory path is not available", func(t *testing.T) {
			// Create a fake $HOME
			home, err := os.MkdirTemp("", "secretless_test")
			assert.NoError(t, err)

			defer func() {
				os.RemoveAll(home)
			}()

			os.Setenv("HOME", home)

			// Override shared memory path
			tempFactory := NewCustomTempFactory("", "doesnotexist")
			assert.Contains(t, tempFactory.path, home)
		})

		t.Run("tries using os.TempDir as last resort", func(t *testing.T) {
			// Override shared memory path
			tempFactory := NewCustomTempFactory("", "doesnotexist")

			assert.Equal(t, os.TempDir(), tempFactory.path)
		})
	})
}
