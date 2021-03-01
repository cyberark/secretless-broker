package mssqltest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/test/connector/tcp/mssql/client"
	mssql "github.com/denisenkom/go-mssqldb"
)

// testClientParams maps a dbClientExecutor to the params that it sends over the wire.
// e.g. sqlcmd sends an applicationName of "SQLCMD" by default.
type testClientParams struct {
	runQuery        client.RunQuery
	applicationName string
	serverName      func(server string, port string) string
}

var sqlcmdParamsTestClient = testClientParams{
	runQuery:        client.SqlcmdExec,
	applicationName: "SQLCMD",
	serverName: func(server string, port string) string {
		return fmt.Sprintf(
			"%s,%s",
			server,
			port,
		)
	},
}

var gomssqlParamsTestClient = testClientParams{
	runQuery:        client.GomssqlExec,
	applicationName: "go-mssqldb",
	serverName: func(server string, port string) string {
		return server
	},
}

var pythonODBCParamsTestClient = testClientParams{
	runQuery:        client.PythonODBCExec,
	applicationName: "python3.7",
	serverName: func(server string, port string) string {
		return fmt.Sprintf(
			"%s,%s",
			server,
			port,
		)
	},
}

type clientParamsTestCase struct {
	description string
	readonly    bool
	testClient  testClientParams
}

func TestClientParams(t *testing.T) {
	testCases := []clientParamsTestCase{
		{
			description: "sqlcmd: client params are propagated to the server",
			readonly:    false,
			testClient:  sqlcmdParamsTestClient,
		},
		{
			description: "go-mssqldb: client params are propagated to the server",
			readonly:    false,
			testClient:  gomssqlParamsTestClient,
		},
		{
			description: "pythonODBC: client params are propagated to the server",
			readonly:    false,
			testClient:  pythonODBCParamsTestClient,
		},
		{
			description: "sqlcmd: readonly application intent is propagated",
			readonly:    true,
			testClient:  sqlcmdParamsTestClient,
		},
		{
			description: "go-mssqldb: readonly application intent is propagated",
			readonly:    true,
			testClient:  gomssqlParamsTestClient,
		},
		{
			description: "pythonODBC: readonly application intent is propagated",
			readonly:    true,
			testClient:  pythonODBCParamsTestClient,
		},
	}

	// Create a single mock target to be used for all the test cases
	mt, err := newMockTarget("0")
	if !assert.NoError(t, err) {
		return
	}
	defer mt.close()

	for _, tc := range testCases {
		t.Run(
			tc.description,
			func(t *testing.T) {
				// 0. Setup expectations
				expectedUsername := "someuser"
				expectedPassword := "somepassword"
				expectedAppname := tc.testClient.applicationName
				expectedDatabase := "random"

				// 1. Create a client request
				clientRequest := clientRequest{
					database: expectedDatabase,
					readOnly: tc.readonly,
					// Ensure query is empty since this request will go to a mock server
					// which is incapable of handling queries.
					query: "",
				}

				// 2. Proxy the client request through a Secretless service to a target
				// service mock using the provided credentials. The goal here is to
				// capture the packets the server receive from an actual client request
				// and ensure that parameters are propagated.
				capture, port, err := clientRequest.proxyToMock(
					tc.testClient.runQuery,
					map[string][]byte{
						"username": []byte(expectedUsername),
						"password": []byte(expectedPassword),
						"sslmode":  []byte("disable"),
					},
					mt,
				)
				if !assert.NoError(t, err) {
					return
				}
				// It is only at this point that we know the Secretless port, and are able
				// to determine the expected server.
				expectedServer := tc.testClient.serverName("127.0.0.1", port)

				// 3. Assert on the login request packet captured by the mock target
				// server to ensure that params are propagated from client to target
				// server through Secretless

				assert.Equal(t, capture.loginRequest.UserName, expectedUsername)
				// Expected password needs to be mangled to match how it is transported
				// to the server
				assert.Equal(t,
					mssql.ManglePassword(expectedPassword),
					[]byte(capture.loginRequest.Password),
				)
				assert.Equal(t, expectedDatabase, capture.loginRequest.Database)
				assert.Equal(t, expectedServer, capture.loginRequest.ServerName)
				assert.Equal(t, expectedAppname, capture.loginRequest.AppName)

				// Conditionally assert on application intent
				if tc.readonly {
					assert.NotEqual(t, int(capture.loginRequest.TypeFlags&32), 0)
				} else {
					assert.Equal(t, int(capture.loginRequest.TypeFlags&32), 0)
				}
			},
		)
	}
}
