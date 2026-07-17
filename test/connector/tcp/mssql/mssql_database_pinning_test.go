package mssqltest

import (
	"testing"

	"github.com/cyberark/secretless-broker/test/connector/tcp/mssql/client"
	"github.com/cyberark/secretless-broker/test/util/testutil"
	"github.com/stretchr/testify/assert"
)

const pinnedMSSQLServicePort = "2225"

func TestDatabasePinning(t *testing.T) {
	testCases := []struct {
		description string
		runQuery    client.RunQuery
	}{
		{"sqlcmd", client.SqlcmdExec},
		{"go-mssql", client.GomssqlExec},
		{"python-ODBC", client.PythonODBCExec},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			t.Run("operator-configured database overrides client-supplied database", func(t *testing.T) {
				// Client requests "different_db"; broker pins to "tempdb" so SELECT DB_NAME()
				// must return the pinned one.
				cfg := client.Config{
					Host:     testutil.SecretlessHost,
					Port:     pinnedMSSQLServicePort,
					Username: "dummy",
					Password: "dummy",
					Database: "different_db",
				}

				out, err := tc.runQuery(cfg, "SELECT DB_NAME() AS [Current Database];")

				if !assert.NoError(t, err) {
					return
				}
				assert.Contains(t, out, "tempdb")
			})
		})
	}
}
