package mssqltest

import (
	"fmt"

	"github.com/cyberark/secretless-broker/test/connector/tcp/mssql/client"
	"github.com/cyberark/secretless-broker/test/util/testutil"
)

// defaultSecretlessDbConfig returns a dbClientConfig that points to the MSSQL Secretless
// service that provides the happy path.
func defaultSecretlessDbConfig() client.Config {
	return client.Config{
		Host:     testutil.SecretlessHost,
		Port:     fmt.Sprintf("%d", testutil.SecretlessPort),
		Username: "dummy",
		Password: "dummy",
	}
}

// envCfg contains the configuration information for the database test instances.
// Information such as host, port, username, password. This information is passed to the
// the testing code as environment variables. See #NewDbConfigFromEnv for additional
// information.
var envCfg = testutil.NewDbConfigFromEnv()
