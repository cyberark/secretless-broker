package mssqltest

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/test/connector/tcp/mssql/client"
)

const encryptionOptionQuery = `
SELECT encrypt_option
FROM sys.dm_exec_connections
WHERE session_id=@@SPID
`

type tlsTestCase struct {
	description string
	credentials map[string][]byte
	assertion   func(t *testing.T, out string, port string, err error)
}

func getServerCert() []byte {
	cert, err := ioutil.ReadFile("./certs/server-cert.pem")
	if err != nil {
		panic("unable to read server certificate")
	}
	return cert
}

var tlsTestCases = []tlsTestCase{
	{
		description: "sslmode=disable",
		credentials: map[string][]byte{
			"sslmode":  []byte("disable"),
			"username": []byte(envCfg.User),
			"password": []byte(envCfg.Password),
			"host":     []byte(envCfg.HostWithTLS),
			"port":     []byte(envCfg.Port),
		},
		assertion: func(t *testing.T, out string, port string, err error) {
			if !assert.NoError(t, err) {
				return
			}

			assert.Contains(t, out, "FALSE")
		},
	},
	{
		description: "sslmode=require",
		credentials: map[string][]byte{
			"sslmode":  []byte("require"),
			"username": []byte(envCfg.User),
			"password": []byte(envCfg.Password),
			"host":     []byte(envCfg.HostWithTLS),
			"port":     []byte(envCfg.Port),
		},
		assertion: func(t *testing.T, out string, port string, err error) {
			if !assert.NoError(t, err) {
				return
			}

			assert.Contains(t, out, "TRUE")
		},
	},
	{
		description: "sslmode=verify-ca: self-signed no sslrootcert",
		credentials: map[string][]byte{
			"sslmode":  []byte("verify-ca"),
			"username": []byte(envCfg.User),
			"password": []byte(envCfg.Password),
			"host":     []byte(envCfg.HostWithTLS),
			"port":     []byte(envCfg.Port),
		},
		assertion: func(t *testing.T, out string, port string, err error) {
			assert.Error(t, err)
		},
	},
	{
		description: "sslmode=verify-ca: self-signed and sslrootcert",
		credentials: map[string][]byte{
			"sslmode":     []byte("verify-ca"),
			"sslrootcert": getServerCert(),
			"username":    []byte(envCfg.User),
			"password":    []byte(envCfg.Password),
			"host":        []byte(envCfg.HostWithTLS),
			"port":        []byte(envCfg.Port),
		},
		assertion: func(t *testing.T, out string, port string, err error) {
			if !assert.NoError(t, err) {
				return
			}

			assert.Contains(t, out, "TRUE")
		},
	},
	{
		description: "sslmode=verify-full: self-signed and no sslrootcert",
		credentials: map[string][]byte{
			"sslmode":  []byte("verify-full"),
			"username": []byte(envCfg.User),
			"password": []byte(envCfg.Password),
			"host":     []byte(envCfg.HostWithTLS),
			"port":     []byte(envCfg.Port),
		},
		assertion: func(t *testing.T, out string, port string, err error) {
			assert.Error(t, err)
		},
	},
	{
		description: "sslmode=verify-full: sslrootcert but hostname mismatch",
		credentials: map[string][]byte{
			"sslmode":     []byte("verify-full"),
			"sslrootcert": getServerCert(),
			"username":    []byte(envCfg.User),
			"password":    []byte(envCfg.Password),
			"host":        []byte(envCfg.HostWithTLS),
			"port":        []byte(envCfg.Port),
		},
		assertion: func(t *testing.T, out string, port string, err error) {
			assert.Error(t, err)
		},
	},
	{
		description: "sslmode=verify-full: sslrootcert and sslhost",
		credentials: map[string][]byte{
			"sslmode":     []byte("verify-full"),
			"sslrootcert": getServerCert(),
			"sslhost":     []byte("mismatchedhost"),
			"username":    []byte(envCfg.User),
			"password":    []byte(envCfg.Password),
			"host":        []byte(envCfg.HostWithTLS),
			"port":        []byte(envCfg.Port),
		},
		assertion: func(t *testing.T, out string, port string, err error) {
			if !assert.NoError(t, err) {
				return
			}

			assert.Contains(t, out, "TRUE")
		},
	},
}

func TestTLS(t *testing.T) {
	for _, testCase := range tlsTestCases {
		t.Run(testCase.description, func(t *testing.T) {
			// Specify client request
			clientReq := clientRequest{
				database: "tempdb",
				readOnly: false,
				query:    encryptionOptionQuery,
			}

			// Proxy Request through Secretless
			out, port, err := clientReq.proxyViaSecretless(
				client.SqlcmdExec,    // Use SQLCMD client
				testCase.credentials, // Credentials from test case
			)

			if !assert.NotNil(t, testCase.assertion) {
				return
			}

			testCase.assertion(t, out, port, err)
		})
	}
}
