package mssqltest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/test/connector/tcp/mssql/client"
	"github.com/cyberark/secretless-broker/test/util/testutil"
)

func TestMSSQLConnector(t *testing.T) {
	testCases := []struct {
		description string
		runQuery    client.RunQuery
	}{
		{
			"python-ODBC",
			client.PythonODBCExec,
		},
		{
			"java-JDBC",
			client.JavaJDBCExec,
		},
		{
			"go-mssql",
			client.GomssqlExec,
		},
		{
			"sqlcmd",
			client.SqlcmdExec,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			RunConnectivityTests(t, testCase.runQuery)
		})
	}
}

func RunConnectivityTests(t *testing.T, runQuery client.RunQuery) {
	t.Run("Can connect to MSSQL through Secretless", func(t *testing.T) {
		testInt := "1+1"
		testString := "abc"

		// Execute query
		out, err := runQuery(
			defaultSecretlessDbConfig(),
			fmt.Sprintf("SELECT %s AS sum, '%s' AS str", testInt, testString),
		)

		// Test the returned values
		if !assert.NoError(t, err) {
			return
		}

		assert.Contains(t, out, "2")
		assert.Contains(t, out, testString)
	})

	t.Run("Cannot connect directly to MSSQL", func(t *testing.T) {
		// Set Host and Port to $DB_HOST_TLS and $DB_PORT environment
		// variables, respectively.
		envCfg := testutil.NewDbConfigFromEnv()

		cfg := defaultSecretlessDbConfig()
		cfg.Host = envCfg.HostWithTLS
		cfg.Port = envCfg.Port

		// This is for local testing. Locally, Secretless and and the target service
		// are exposed on 127.0.0.1 via port mappings
		if testutil.SecretlessHost == "127.0.0.1" {
			cfg.Host = "127.0.0.1"
		}

		// Execute query
		_, err := runQuery(cfg, "")

		// Test the returned values
		if !assert.Error(t, err, "direct db connection should error") {
			return
		}
		assert.Contains(t, err.Error(), "Login failed")
	})

	t.Run("Passes valid database name to MSSQL through Secretless", func(t *testing.T) {
		cfg := defaultSecretlessDbConfig()
		// Existing database name, see
		// https://docs.microsoft.com/en-us/sql/relational-databases/databases/tempdb-database?view=sql-server-ver15
		cfg.Database = "tempdb"

		// Execute query
		out, err := runQuery(
			cfg,
			"SELECT DB_NAME() AS [Current Database];", // returns current database
		)

		// Test the returned values
		if !assert.NoError(t, err, "valid db should not error") {
			return
		}
		assert.Contains(t, out, "tempdb")
	})

	t.Run("Passes invalid database name to MSSQL through Secretless", func(t *testing.T) {
		cfg := defaultSecretlessDbConfig()
		// Non-existent database name
		cfg.Database = "meow"

		// Execute query
		_, err := runQuery(
			cfg,
			"",
		)
		// Test the returned values
		if !assert.Error(t, err, "invalid db should error") {
			return
		}
		assert.Contains(t, err.Error(), "Cannot open database")
	})

}
