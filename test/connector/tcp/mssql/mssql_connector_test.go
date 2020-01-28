package mssqltest

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/test/util/testutil"
)

func TestMSSQLConnector(t *testing.T) {
	t.Run("go-mssql", func(t *testing.T) {
		RunTests(t, gomssqlExec)
	})

	t.Run("sqlcmd", func(t *testing.T) {
		RunTests(t, sqlcmdExec)
	})
}

func RunTests(t *testing.T, queryExec dbQueryExecutor) {
	t.Run("Can connect to MSSQL through Secretless", func(t *testing.T) {
		testInt := "1+1"
		testString := "abc"

		// Execute Query
		out, err := queryExec(
			defaultSecretlessDbConfig(),
			fmt.Sprintf("select %s, '%s'", testInt, testString),
		)

		// Test the returned values
		assert.NoError(t, err)

		assert.Contains(t, out, "2")
		assert.Contains(t, out, testString)
	})

	t.Run("Cannot connect directly to MSSQL", func(t *testing.T) {
		// Set Host and Port to $DB_HOST_TLS and $DB_PORT environment
		// variables, respectively.
		envCfg := testutil.NewDbConfigFromEnv()
		cfg := defaultSecretlessDbConfig()
		cfg.Host = envCfg.HostWithTLS
		var err error
		cfg.Port, err = strconv.Atoi(envCfg.Port)
		assert.NoError(t, err)

		// This is for local testing. Locally, Secretless and and the target service
		// are exposed on 127.0.0.1 via port mappings
		if testutil.SecretlessHost == "127.0.0.1" {
			cfg.Host = "127.0.0.1"
		}

		// Execute Query
		_, err = queryExec(
			cfg,
			"",
		)

		// Test the returned values
		assert.Error(t, err, "direct db connection should error")
		assert.Contains(t, err.Error(), "Login failed")
	})

	t.Run("Passes valid database name to MSSQL through Secretless", func(t *testing.T) {
		cfg := defaultSecretlessDbConfig()
		// existing database name, see
		// https://docs.microsoft.com/en-us/sql/relational-databases/databases/tempdb-database?view=sql-server-ver15
		cfg.Database = "tempdb"

		// Execute Query
		out, err := queryExec(
			cfg,
			"SELECT DB_NAME() AS [Current Database];", // returns current database
		)

		// Test the returned values
		assert.NoError(t, err, "valid db should not error")
		assert.Contains(t, out, "tempdb")
	})

	// TODO: The following test depends upon the changes that will be
	// implemented in PR #1111. For now, this is disabled.
	//t.Run("Passes invalid database name to MSSQL through Secretless", func(t *testing.T) {
	//	cfg := defaultSecretlessDbConfig()
	//	// non-existent database name
	//	cfg.Database = "meow"

	//	// Execute Query
	//	_, err := queryExec(
	//		cfg,
	//		"",
	//	)
	//	// Test the returned values
	//	assert.Error(t, err, "invalid db should error")
	//	if err == nil {
	//		return
	//	}
	//	assert.Contains(t, err.Error(), "Cannot open database")
	//})

}
