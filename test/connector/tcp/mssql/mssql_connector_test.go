package mssqltest

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/test/util/testutil"
	mssql "github.com/denisenkom/go-mssqldb"
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

	t.Run("Passes invalid database name to MSSQL through Secretless", func(t *testing.T) {
		cfg := defaultSecretlessDbConfig()
		// non-existent database name
		cfg.Database = "meow"

		// Execute Query
		_, err := queryExec(
			cfg,
			"",
		)

		// Test the returned values
		assert.Error(t, err, "invalid db should error")
		assert.Contains(t, err.Error(), "Cannot open database")
	})

}

const mockServerPort = 2224
type testClientParams struct {
	queryExec dbQueryExecutor
	applicationName string
	serverName func(server string, port int) string
}

func TestClientParams(t *testing.T) {
	// Setup mock-server listener
	ln, err := net.Listen("tcp", ":1434")
	defer func() {
		_ = ln.Close()
	}()

	if err != nil {
		panic(err)
	}

	sqlcmdParamsTestClient := testClientParams{
		queryExec:        sqlcmdExec,
		applicationName: "SQLCMD",
		serverName: func(server string, port int) string {
			return fmt.Sprintf(
				"%s,%d",
				server,
				port,
			)
		},
	}
	gomssqlParamsTestClient := testClientParams{
		queryExec:        gomssqlExec,
		applicationName: "go-mssqldb",
		serverName: func(server string, port int) string {
			return server
		},
	}

	t.Run(
		"sqlcmd: client params are propagated to the server",
		func(t *testing.T) {
			propagateParams(
				t,
				false,
				ln,
				sqlcmdParamsTestClient,
			)
		},
	)

	t.Run(
		"go-mssqldb: client params are propagated to the server",
		func(t *testing.T) {
			propagateParams(
				t,
				false,
				ln,
				gomssqlParamsTestClient,
			)
		},
	)

	t.Run(
		"sqlcmd: readonly application intent is propagated",
		func(t *testing.T) {
			propagateParams(
				t,
				true,
				ln,
				sqlcmdParamsTestClient,
			)
		},
	)

	t.Run(
		"go-mssqldb: readonly application intent is propagated",
		func(t *testing.T) {
			propagateParams(
				t,
				true,
				ln,
				gomssqlParamsTestClient,
			)
		},
	)
}

func propagateParams(
	t *testing.T,
	readonly bool,
	listener net.Listener,
	testClient testClientParams,
) {
	// these are credential values injected by secretless
	expectedUsername := "expected-user"
	expectedPassword := "expected-password"

	// this is the database name the client passes in
	expectedDatabase := "expected-database"

	// this is configurable because each client has it's own application name
	expectedAppname := testClient.applicationName
	expectedServer := testClient.serverName(testutil.SecretlessHost, mockServerPort)

	// 1. Make client connection

	go func() {
		// we don't actually care about the response
		_, _ = testClient.queryExec(
			dbConfig{
				Host:     testutil.SecretlessHost,
				Port:     mockServerPort,
				Username: "dummy",
				Password: "dummy",
				Database: expectedDatabase,
				ReadOnly: readonly,
			},
			"",
		)
	}()

	// 2. Accept client connection and carry out login handshake

	clientConnection, err := listener.Accept()
	if err != nil {
		panic(err)
	}

	// Set a deadline so that if things hang then they fail fast
	readWriteDeadline := time.Now().Add(5 * time.Second)
	err = clientConnection.SetDeadline(readWriteDeadline)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = clientConnection.Close()
	}()
	if err != nil {
		panic(err)
	}

	// read prelogin request
	preloginRequest, err := mssql.ReadPreloginRequest(clientConnection)
	if err != nil {
		panic(err)
	}
	// write prelogin response
	// ensuring no TLS support
	preloginResponse := preloginRequest
	preloginResponse[mssql.PreloginVERSION] = []byte{0x0e, 0x00, 0x0c, 0xa6, 0x00, 0x00}
	preloginResponse[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}
	err = mssql.WritePreloginResponse(clientConnection, preloginResponse)
	if err != nil {
		panic(err)
	}

	// write login request
	loginRequest, err := mssql.ReadLoginRequest(clientConnection)
	if err != nil {
		panic(err)
	}
	// write a dummy successful login response
	loginResponse := &mssql.LoginResponse{}
	loginResponse.ProgName = "test"
	loginResponse.TDSVersion = 0x730A0003
	loginResponse.Interface = 27
	err = mssql.WriteLoginResponse(clientConnection, loginResponse)
	if err != nil {
		panic(err)
	}

	// 3. Test the captured login request

	assert.Equal(t, loginRequest.UserName, expectedUsername)
	// expected password needs to be mangled to match how it is transported to the server
	assert.Equal(t, mssql.ManglePassword(expectedPassword), []byte(loginRequest.Password))
	assert.Equal(t, expectedDatabase, loginRequest.Database)
	assert.Equal(t, expectedServer, loginRequest.ServerName)
	assert.Equal(t, expectedAppname, loginRequest.AppName)

	// conditionally assert on application intent
	if readonly {
		assert.NotEqual(t, int(loginRequest.TypeFlags & 32), 0)
	} else {
		assert.Equal(t, int(loginRequest.TypeFlags & 32), 0)
	}
}