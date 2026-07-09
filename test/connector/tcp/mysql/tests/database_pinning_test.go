package tests

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/cyberark/secretless-broker/test/util/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabasePinning verifies that when `database` is set in the broker's
// credentials, the client-supplied database is ignored and the operator-configured
// one is used instead.
func TestDatabasePinning(t *testing.T) {
	t.Run("operator-configured database is used regardless of client request", func(t *testing.T) {
		// Client requests "different_db"; broker pins to "testdb" so SELECT DATABASE()
		// must return the pinned one.
		args := []string{
			fmt.Sprintf("--host=%s", testutil.SecretlessHost),
			fmt.Sprintf("--port=%s", testutil.PinnedDatabaseServicePort),
			"--skip-ssl",
			"--database=different_db",
			"-e", "SELECT DATABASE();",
		}

		out, err := exec.Command("mysql", args...).CombinedOutput()

		require.NoError(t, err)
		assert.Contains(t, string(out), "testdb")
	})
}
