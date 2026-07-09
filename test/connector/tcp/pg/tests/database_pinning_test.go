package tests

import (
	"testing"

	"github.com/cyberark/secretless-broker/test/util/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabasePinning verifies that when `database` is set in the broker's
// credentials, the client-supplied database is ignored and the operator-configured
// one is used instead.
func TestDatabasePinning(t *testing.T) {
	runQuery := genQueryRunner(testutil.SecretlessHost, testutil.PinnedDatabaseServicePort)

	t.Run("operator-configured database is used regardless of client request", func(t *testing.T) {
		// Client requests "different_db"; broker pins to "postgres" so the query must succeed
		// and current_database() must return the pinned one.
		out, err := runQuery(nil, []string{"dbname=different_db"}, "SELECT current_database();")

		require.NoError(t, err)
		assert.Contains(t, out, "postgres")
	})
}
