package tests

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/cyberark/secretless-broker/test/util/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClientOptions(t *testing.T) {
	ac := testutil.AbstractConfiguration{
		SocketType:     testutil.TCP,
		TLSSetting:     testutil.TLS,
		SSLMode:        testutil.Default,
		RootCertStatus: testutil.Undefined,
	}
	l := testutil.FindLiveConfiguration(ac)

	options := []string{"dbname=postgres"}

	t.Run("dbname", func(t *testing.T) {
		t.Run("exisiting", func(t *testing.T) {
			output, err := runQuery(l.Host(), l.ToPortString(), nil, options, "SELECT current_database();")

			assert.NoError(t, err)
			assert.Contains(t, output, "postgres")
		})

		t.Run("not found", func(t *testing.T) {
			output, err := runQuery(l.Host(), l.ToPortString(), nil, []string{"dbname=notfound"}, "SELECT current_database();")

			assert.Error(t, err)
			assert.Contains(t, output, `database "notfound" does not exist`)
		})
	})

	t.Run("client_encoding", func(t *testing.T) {
		genEnvs := func(encoding string) []string {
			return []string{
				fmt.Sprintf("PGCLIENTENCODING=%s", encoding),
			}
		}

		t.Run("client encoding [smoke test]", func(t *testing.T) {
			encoding := "latin1"
			query := fmt.Sprintf("select value from test.encodings where encoding='%s'", encoding)

			output, err := runQuery(l.Host(), l.ToPortString(), genEnvs(encoding), options, query)

			assert.NoError(t, err)
			assert.Contains(t, output, "tést")
		})

		t.Run("defaults to utf8 when not set", func(t *testing.T) {
			output, err := runQuery(l.Host(), l.ToPortString(), nil, options, "SHOW client_encoding;")

			assert.NoError(t, err)
			assert.Contains(t, output, "UTF8")
		})

		t.Run("propagates to target server", func(t *testing.T) {
			output, err := runQuery(l.Host(), l.ToPortString(), genEnvs("euc-jp"), options, "SHOW client_encoding;")

			assert.NoError(t, err)
			assert.Contains(t, output, "EUC_JP")
		})
	})
}

func runQuery(host string, port string, envs []string, options []string, query string) (string, error) {
	args := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--port=%s", port),
		"-c", query,
		strings.Join(append([]string{"sslmode=disable"}, options...), " "),
	}

	c := exec.Command("psql", args...)
	c.Env = append(c.Env, envs...)
	cmdOut, err := c.CombinedOutput()

	return string(cmdOut), err
}
