package v1

import (
	"os"
	"net"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestListener(t *testing.T) {
	socketFile := "/sock/test"
	if _, err := os.Stat(socketFile); os.IsNotExist(err) {
		socketFile = "./sock"
	}

	netListener, _ := net.Listen("unix", socketFile)
	listener := NewBaseListener(ListenerOptions{
		NetListener: netListener,
	})

	Convey("BaseListener shuts down cleanly without errors", t, func() {
		// First, check that the socket file exists
		_, err := os.Stat(socketFile)
		So(err, ShouldBeNil)

		err = listener.Shutdown()
		So(err, ShouldBeNil)

		Convey("and its socket file is removed", func() {
			_, err := os.Stat(socketFile)
			So(err, ShouldNotBeNil)
			So(os.IsNotExist(err), ShouldBeTrue)
		})
	})
}