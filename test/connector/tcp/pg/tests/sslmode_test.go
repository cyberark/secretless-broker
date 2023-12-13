package tests

import (
	"os/exec"
	"strings"
	"testing"

	. "github.com/cyberark/secretless-broker/test/util/testutil"
	"github.com/stretchr/testify/assert"
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

	t.Run("sslmode=require", func(t *testing.T) {
		cmdOut, _ := exec.Command("psql", append(args, strings.Join(append(connectionParams, "sslmode=require"), " "))...).CombinedOutput()
		assert.Contains(t, string(cmdOut), "psql: error: server does not support SSL, but SSL was required")
	})

	t.Run("sslmode=prefer", func(t *testing.T) {
		cmdOut, _ := exec.Command("psql", append(args, strings.Join(append(connectionParams, "sslmode=prefer"), " "))...).CombinedOutput()
		assert.Contains(t, string(cmdOut), "count")
	})

}
