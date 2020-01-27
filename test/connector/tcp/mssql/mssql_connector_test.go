package mssqltest

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/test/util/testutil"
	mssql "github.com/denisenkom/go-mssqldb"
)

var _ = Describe("MSSQL Connector Test", func() {

	TestClients := []struct {
		appName string
		queryExec  dbQueryExecutor
		serverName func(server string, port int) string
	}{
		{
			appName: "go-mssql",
			queryExec: gomssqlExec,
			serverName: func(server string, port int) string {
				return fmt.Sprintf(
					"%s,%d",
					server,
					port,
				)
			},
		},
		{
			appName: "sqlcmd",
			queryExec: sqlcmdExec,
			serverName: func(server string, port int) string {
				return server
			},
		},
	}

	for _, tc := range TestClients {

		Describe(fmt.Sprintf("connectivity tests with %s", tc.appName), func() {
			const (
				selectNum    = "1+1"
				selectString = "1+1"
			)
			var (
				// These variables represent test configuration or state
				cfg      dbConfig
				query    string
				queryOut string
				queryErr error
				err      error
			)

			BeforeEach(func() {
				// Set defaults. May be overriden by test cases.
				cfg = defaultSecretlessDbConfig()
				query = fmt.Sprintf("select %s, '%s'", selectNum, selectString)
			})

			JustBeforeEach(func() {
				// Run the query
				queryOut, queryErr = tc.queryExec(cfg, query)
			})

			Describe("can connect to MSSQL through Secretless", func() {

				// Query will have been run in JustBeforeEach block

				It("should not error", func() {
					Expect(queryErr).NotTo(HaveOccurred())
				})

				It("should return expected content", func() {
					Expect(queryOut).To(ContainSubstring("2"))
					Expect(queryOut).To(ContainSubstring(selectString))
				})
			})

			Describe("cannot connect directly to MSSQL", func() {
				BeforeEach(func() {
					// Set Host and Port to $DB_HOST_TLS and $DB_PORT
					// environment variables, respectively.
					envCfg := testutil.NewDbConfigFromEnv()
					cfg.Host = envCfg.HostWithTLS
					cfg.Port, err = strconv.Atoi(envCfg.Port)
					Expect(err).NotTo(HaveOccurred())

					// This is for local testing. Locally, Secretless
					// and and the target service are exposed on
					// 127.0.0.1 via port mappings
					if testutil.SecretlessHost == "127.0.0.1" {
						cfg.Host = "127.0.0.1"
					}
				})

				It("should get login failed error from MSSQL server", func() {
					// Test the returned values
					Expect(queryErr).To(HaveOccurred())
					Expect(queryErr.Error()).To(ContainSubstring("Login failed"))
				})
			})

			Describe("passes valid database name to MSSQL through Secretless", func() {
				BeforeEach(func() {
					// existing database name, see
					// https://docs.microsoft.com/en-us/sql/relational-databases/databases/tempdb-database?view=sql-server-ver15
					cfg.Database = "tempdb"
					// Set query to return current database
					query = "SELECT DB_NAME() AS [Current Database];"
				})

				It("should return a valid database name", func() {
					// Test the returned values
					Expect(queryErr).ToNot(HaveOccurred())
					Expect(queryOut).To(ContainSubstring("tempdb"))
				})
			})

			Describe("passes invalid database name to MSSQL through Secretless", func() {
				BeforeEach(func() {
					// non-existent database name
					cfg.Database = "meow"
				})

				It("should receive a 'Cannot open database' error", func() {
					Expect(queryErr).To(HaveOccurred())
					Expect(queryErr.Error()).To(ContainSubstring("Cannot open database"))
				})
			})
		})

		description := fmt.Sprintf("parameter propagation tests with %s", tc.appName)
		Describe(description, func () {
			const (
				// these are credential values injected by secretless
				expUsername = "expected-user"
				expPassword = "expected-password"

				// this is the database name the client passes in
				expDatabase = "expected-database"
			)
			var (
				expAppname string
				expServer string
			)
			BeforeEach(func() {
				expAppname := tc.appName
				expServer := tc.serverName(
					testutil.SecretlessHost,
					mockServerPort)
			})

			It("should set up a mock server listening on port 1434", func() {
				// Setup mock-server listener
				ln, err := net.Listen("tcp", ":1434")
				defer func() {
					_ = ln.Close()
				}()
				Expect(err).ToNot(HaveOccurred())
			})

			for _, readonly := range []bool{true, false} {

				Context(fmt.Sprintf("test propagation with readonly=%v", readonly), func() {

					By("making a connection to Secretless", func() {
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
					})

					By("accepting connection from secretless
						It("accept
							It("
							BeforeEach(
						Before
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




				
			func(readonly bool) {

			DescribeTable("k
			var (
				readonly bool
			)

			propagateParameters := func() {

			Context("while running a mock SQL server on port 1434", func() {
				It("should be able to listen on port 1434", func() {
					// Setup mock-server listener
					ln, err := net.Listen("tcp", ":1434")
					defer func() {
						_ = ln.Close()
					}()
					Expect(err).ToNot(HaveOccurred())
				}

				Describe("client params are propagated to the server", func() {


	if err != nil {
		panic(err)
	}

})

const mockServerPort = 2224

type testClientParams struct {
	queryExec       dbQueryExecutor
	applicationName string
	serverName      func(server string, port int) string
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
		queryExec:       sqlcmdExec,
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
		queryExec:       gomssqlExec,
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
		assert.NotEqual(t, int(loginRequest.TypeFlags&32), 0)
	} else {
		assert.Equal(t, int(loginRequest.TypeFlags&32), 0)
	}
}
