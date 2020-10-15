package command

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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

func assertMissingFile(f string) {
	_, err := os.Stat(f)
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldContainSubstring, "no such file or directory")
}

func assertFileContents(f string, expectedValue string) {
	actualContent, err := ioutil.ReadFile(f)
	So(err, ShouldBeNil)
	So(expectedValue, ShouldEqual, string(actualContent))
}

func TestTempFactory_Cleanup(t *testing.T) {
	Convey("Cleanup deletes all temp files", t, func() {
		tempFactory := NewCustomTempFactory("", "non-existent")

		f1, err := tempFactory.Push("meow")
		So(err, ShouldBeNil)
		f2, err := tempFactory.Push("moo")
		So(err, ShouldBeNil)

		assertFileContents(f1, "meow")
		assertFileContents(f2, "moo")

		tempFactory.Cleanup()

		assertMissingFile(f1)
		assertMissingFile(f2)
	})
}

func TestTempFactory_Push(t *testing.T) {
	Convey("Push creates temp file", t, func() {
		tempFactory := NewTempFactory("")
		defer tempFactory.Cleanup()

		f, err := tempFactory.Push("moo")
		So(err, ShouldBeNil)
		assertFileContents(f, "moo")
	})

	Convey("Push reports errors", t, func() {
		tempFactory := NewTempFactory("dir-not-found")
		defer tempFactory.Cleanup()

		_, err := tempFactory.Push("moo")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "no such file or directory")
	})
}

func TestTempFactory_NewTempFactory(t *testing.T) {
	Convey("Uses constructor arg path if provided", t, func() {
		tempFactory := NewTempFactory("somedir")
		defer tempFactory.Cleanup()

		So(tempFactory, ShouldResemble, TempFactory{
			files: []string(nil),
			path:  "somedir",
		})
	})

	Convey("When constructor path is not provided", t, func() {
		env := clearEnv()
		defer env.restoreEnv()

		Convey("tries using shared memory path first", func() {
			tempFactory := NewTempFactory("")

			_, err := os.Stat("/dev/shm")
			if os.IsNotExist(err) {
				return
			}

			So(tempFactory, ShouldResemble, TempFactory{
				files: []string(nil),
				path:  "/dev/shm",
			})
		})

		Convey("tries using homedir prefix if shared memory path is not available", func() {
			// Create a fake $HOME
			home, err := ioutil.TempDir("", "secretless_test")
			So(err, ShouldBeNil)

			defer func() {
				os.RemoveAll(home)
			}()

			os.Setenv("HOME", home)

			// Override shared memory path
			tempFactory := NewCustomTempFactory("", "doesnotexist")
			So(tempFactory.path, ShouldStartWith, home)
		})

		Convey("tries using os.TempDir as last resort", func() {
			// Override shared memory path
			tempFactory := NewCustomTempFactory("", "doesnotexist")

			So(tempFactory.path, ShouldEqual, os.TempDir())
		})
	})
}
