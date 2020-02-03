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
	t.Run("python-ODBC", func(t *testing.T) {
		RunConnectivityTests(t, pythonODBCExec)
	})

	t.Run("java-JDBC", func(t *testing.T) {
		RunConnectivityTests(t, javaJDBCExec)
	})

	t.Run("go-mssql", func(t *testing.T) {
		RunConnectivityTests(t, gomssqlExec)
	})

	t.Run("sqlcmd", func(t *testing.T) {
		RunConnectivityTests(t, sqlcmdExec)
	})
}

func RunConnectivityTests(t *testing.T, queryExec dbQueryExecutor) {
	t.Run("Can connect to MSSQL through Secretless", func(t *testing.T) {
		testInt := "1+1"
		testString := "abc"

		// Execute Query
		out, err := queryExec(
			defaultSecretlessDbConfig(),
			fmt.Sprintf("SELECT %s AS sum, '%s' AS str", testInt, testString),
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
		if err == nil {
			return
		}
		assert.Contains(t, err.Error(), "Cannot open database")
	})

}

const mockServerSecretlessPort = 2224
type testClientParams struct {
	queryExec dbQueryExecutor
	applicationName string
	serverName func(server string, port int) string
}

func TestClientParams(t *testing.T) {
	// Setup mock-server listener
	_ln, err := net.Listen("tcp", ":1434")
	ln := _ln.(*net.TCPListener)
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
	pythonODBCParamsTestClient := testClientParams{
		queryExec:        pythonODBCExec,
		applicationName: "python3.5",
		serverName: func(server string, port int) string {
			return fmt.Sprintf(
				"%s,%d",
				server,
				port,
			)
		},
	}

	type testcase struct {
		description string
		readonly bool
		testClient testClientParams
	}

	testCases := []testcase{
		{
			description: "sqlcmd: client params are propagated to the server",
			readonly: false,
			testClient: sqlcmdParamsTestClient,
		},
		{
			description: "go-mssqldb: client params are propagated to the server",
			readonly: false,
			testClient: gomssqlParamsTestClient,
		},
		{
			description: "pythonODBC: client params are propagated to the server",
			readonly: false,
			testClient: pythonODBCParamsTestClient,
		},
		{
			description: "sqlcmd: readonly application intent is propagated",
			readonly: true,
			testClient: sqlcmdParamsTestClient,
		},
		{
			description: "go-mssqldb: readonly application intent is propagated",
			readonly: true,
			testClient: gomssqlParamsTestClient,
		},
		{
			description: "pythonODBC: readonly application intent is propagated",
			readonly: true,
			testClient: pythonODBCParamsTestClient,
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.description,
			func(t *testing.T) {
				propagateParams(
					t,
					tc.readonly,
					ln,
					tc.testClient,
				)
			},
		)
	}
}

func propagateParams(
	t *testing.T,
	readonly bool,
	listener *net.TCPListener,
	testClient testClientParams,
) {
	// these are credential values injected by secretless
	expectedUsername := "expected-user"
	expectedPassword := "expected-password"

	// this is the database name the client passes in
	expectedDatabase := "expected-database"

	// this is configurable because each client has it's own application name
	expectedAppname := testClient.applicationName
	expectedServer := testClient.serverName(testutil.SecretlessHost, mockServerSecretlessPort)

	// 1. Make client connection

	go func() {
		// we don't actually care about the response
		_, _ = testClient.queryExec(
			dbConfig{
				Host:     testutil.SecretlessHost,
				Port:     mockServerSecretlessPort,
				Username: "dummy",
				Password: "dummy",
				Database: expectedDatabase,
				ReadOnly: readonly,
			},
			"",
		)
	}()

	// 2. Accept client connection and carry out login handshake

	_ = listener.SetDeadline(time.Now().Add(2 * time.Second))
	clientConnection, err := listener.Accept()
	if !assert.NoError(t, err) {
		return
	}

	// Set a deadline so that if things hang then they fail fast
	readWriteDeadline := time.Now().Add(5 * time.Second)
	err = clientConnection.SetDeadline(readWriteDeadline)
	if !assert.NoError(t, err) {
		return
	}

	defer func() {
		_ = clientConnection.Close()
	}()
	if !assert.NoError(t, err) {
		return
	}

	// read prelogin request
	preloginRequest, err := mssql.ReadPreloginRequest(clientConnection)
	if !assert.NoError(t, err) {
		return
	}
	// write prelogin response
	// ensuring no TLS support
	preloginResponse := preloginRequest
	preloginResponse[mssql.PreloginVERSION] = []byte{0x0e, 0x00, 0x0c, 0xa6, 0x00, 0x00}
	preloginResponse[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}
	err = mssql.WritePreloginResponse(clientConnection, preloginResponse)
	if !assert.NoError(t, err) {
		return
	}

	// read login request
	loginRequest, err := mssql.ReadLoginRequest(clientConnection)
	if !assert.NoError(t, err) {
		return
	}
	// write a dummy successful login response
	loginResponse := &mssql.LoginResponse{}
	loginResponse.ProgName = "test"
	loginResponse.TDSVersion = 0x730A0003
	loginResponse.Interface = 27
	err = mssql.WriteLoginResponse(clientConnection, loginResponse)
	if !assert.NoError(t, err) {
		return
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