package mssqltest

import "github.com/cyberark/secretless-broker/test/util/testutil"

func defaultSecretlessDbConfig() dbConfig {
	return dbConfig{
		Host:     testutil.SecretlessHost,
		Port:     testutil.SecretlessPort,
		Username: "dummy",
		Password: "dummy",
	}
}
