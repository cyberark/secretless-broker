package tests

import (
	"os/exec"
	"strings"
	"testing"

	. "github.com/cyberark/secretless-broker/test/util/testutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCustom(t *testing.T) {

	lc := FindLiveConfiguration(
		AbstractConfiguration{
			SocketType: TCP,
			TLSSetting: TLS,
			SSLMode:    Default,
		},
	)

	connectPort := lc.ConnectionPort

	connectionParams := []string{"dbname=postgres", "password=x"}
	args := []string{
		"-c", "select count(*) from test.test",
		"--username", "x",
		"--port", connectPort.ToPortString(),
		"--host", connectPort.Host(),
	}

	Convey("sslmode=require", t, func() {
		cmdOut, _ := exec.Command("psql", append(args, strings.Join(append(connectionParams, "sslmode=require"), " "))...).CombinedOutput()
		So(string(cmdOut), ShouldContainSubstring, "psql: server does not support SSL, but SSL was required")
	})

	Convey("sslmode=prefer", t, func() {
		cmdOut, _ := exec.Command("psql", append(args, strings.Join(append(connectionParams, "sslmode=prefer"), " "))...).CombinedOutput()
		So(string(cmdOut), ShouldContainSubstring, "count")
	})

}
