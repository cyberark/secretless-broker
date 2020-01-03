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
		// Open connection and run test query
		testInt := "1+1"
		testString := "abc"
		out, err := queryExec(
			defaultSecretlessDbConfig(),
			fmt.Sprintf("select %s, '%s'", testInt, testString),
		)

		assert.NoError(t, err)

		// Test the returned values
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

		_, err = queryExec(
			cfg,
			"",
		)

		assert.Contains(t, err.Error(), "Login failed")
	})

	t.Run("Passes valid database name to MSSQL through Secretless", func(t *testing.T) {
		// Open connection and run test query
		cfg := defaultSecretlessDbConfig()
		cfg.Database = "master"
		_, err := queryExec(
			cfg,
			"",
		)

		assert.NoError(t, err, "valid db should not error")
	})

	t.Run("Passes invalid database name to MSSQL through Secretless", func(t *testing.T) {
		cfg := defaultSecretlessDbConfig()
		cfg.Database = "meow"
		_, err := queryExec(
			cfg,
			"",
		)

		assert.Error(t, err, "invalid db should error")
		assert.Contains(t, err.Error(), "Generic SQL Error")
	})

}
