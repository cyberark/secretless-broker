package tests

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/cyberark/secretless-broker/test/util/testutil"
	"github.com/stretchr/testify/assert"
)

func TestEncodings(t *testing.T) {
	ac := testutil.AbstractConfiguration{
		SocketType:     testutil.TCP,
		TLSSetting:     testutil.TLS,
		SSLMode:        testutil.Default,
		RootCertStatus: testutil.Undefined,
	}
	l := testutil.FindLiveConfiguration(ac)

	// Only "latin1" is covered in this test case. To add more simply loop and ensure that "test/connector/tcp/pg/test.sql" is updated with new encodings.
	encoding := "latin1"
	output, err := selectEncodedValue(l.Host(), l.ToPortString(), encoding)
	assert.NoError(t, err)
	assert.Contains(t, output, "tést")
}

func selectEncodedValue(host string, port string, encoding string) (string, error) {
	args := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--port=%s", port),
		"-c", fmt.Sprintf("select value from test.encodings where encoding='%s'", encoding),
		strings.Join([]string{"dbname=postgres sslmode=disable"}, " "),
	}

	c := exec.Command("psql", args...)
	c.Env = append(c.Env, fmt.Sprintf("PGCLIENTENCODING=%s", encoding))
	cmdOut, err := c.CombinedOutput()

	return string(cmdOut), err
}
