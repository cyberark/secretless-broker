package mssql

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/connectiondetails"
)

type args struct {
	credentials map[string][]byte
}

var defaultConnectionDetails = &connectiondetails.ConnectionDetails{
	Username: "herp",
	Password: "derp",
	Host:     "0.0.0.0",
	Port:     "1234",
	Options: map[string]string{
		"encrypt":                "true",
		"trustservercertificate": "true",
	},
}

var emptyConnectionDetails = &connectiondetails.ConnectionDetails{
	Port: defaultMSSQLPort,
	Options: map[string]string{
		"encrypt":                "true",
		"trustservercertificate": "true",
	},
}

// Test case needed for NewConnectionDetails with handler injection

func TestDefaultSSLModeExists(t *testing.T) {
	assert.NotEmpty(t, sslModeToBaseParams[string(defaultSSLMode)])
}

func TestConnectionDetails_HandleSSLOptions(t *testing.T) {
	testCases := []struct {
		description string
		args        args
		expected    map[string]string
	}{
		{
			description: "empty credentials map",
			args: args{
				credentials: map[string][]byte{},
			},
			expected: sslModeToBaseParams[sslModeRequire],
		},
		{
			description: "sslmode:disable",
			args: args{
				credentials: map[string][]byte{
					"sslmode":     []byte("disable"),
					"sslrootcert": []byte("foo"),
				},
			},
			expected: map[string]string{
				"encrypt": "disable",
			},
		},
		{
			description: "sslmode:verify-ca",
			args: args{
				credentials: map[string][]byte{
					"sslmode":     []byte("verify-ca"),
					"sslrootcert": []byte("foo"),
				},
			},
			expected: map[string]string{
				"encrypt":                "true",
				"trustservercertificate": "false",
				"disableverifyhostname":  "true",
				"rawcertificate":         "foo",
			},
		},
		{
			description: "sslmode:verify-full",
			args: args{
				credentials: map[string][]byte{
					"sslmode":     []byte("verify-full"),
					"sslrootcert": []byte("foo"),
				},
			},
			expected: map[string]string{
				"encrypt":                "true",
				"trustservercertificate": "false",
				"disableverifyhostname":  "false",
				"rawcertificate":         "foo",
			},
		},
		{
			description: "sslmode:verify-full with sslhost",
			args: args{
				credentials: map[string][]byte{
					"sslmode":     []byte("verify-full"),
					"sslhost":     []byte("foo.bar"),
					"sslrootcert": []byte("foo"),
				},
			},
			expected: map[string]string{
				"encrypt":                "true",
				"trustservercertificate": "false",
				"disableverifyhostname":  "false",
				"rawcertificate":         "foo",
				"hostnameincertificate":  "foo.bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actualSSLOptions, _ := HandleSSLOptions(tc.args.credentials)

			assert.Equal(t, tc.expected, actualSSLOptions)
		})
	}
}
