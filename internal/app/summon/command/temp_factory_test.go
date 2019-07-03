package command

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TODO: Test TempFactory.Cleanup()
// TODO: Test TempFactory.Push()

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

func TestTempFactory_NewTempFactory(t *testing.T) {
	Convey("Uses constructor arg path if provided", t, func() {
		tempFactory := NewTempFactory("somedir")

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
