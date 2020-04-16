package mssql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type args struct {
	credentials map[string][]byte
}

var defaultConnectionDetails = &ConnectionDetails{
	Username: "herp",
	Password: "derp",
	Host:     "0.0.0.0",
	Port:     1234,
	SSLParams: map[string]string{
		"encrypt":                "true",
		"trustservercertificate": "true",
	},
}

var emptyConnectionDetails = &ConnectionDetails{
	Port: defaultMSSQLPort,
	SSLParams: map[string]string{
		"encrypt":                "true",
		"trustservercertificate": "true",
	},
}

func TestConnectionDetails_Address(t *testing.T) {
	testCases := []struct {
		description string
		fields      *ConnectionDetails
		expected    string
	}{
		{
			description: "expected valid input",
			fields:      defaultConnectionDetails,
			expected:    "0.0.0.0:1234",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			cd := &ConnectionDetails{
				Host:      tc.fields.Host,
				Port:      tc.fields.Port,
				Username:  tc.fields.Username,
				Password:  tc.fields.Password,
				SSLParams: tc.fields.SSLParams,
			}

			assert.Equal(t, tc.expected, cd.address())
		})
	}
}

func TestNewConnectionDetails(t *testing.T) {
	testCases := []struct {
		description string
		args        args
		expected    *ConnectionDetails
	}{
		{
			description: "expected valid credentials",
			args: args{
				credentials: map[string][]byte{
					"username":    []byte("herp"),
					"password":    []byte("derp"),
					"host":        []byte("0.0.0.0"),
					"port":        []byte("1234"),
					"sslmode":     []byte("require"),
					"sslrootcert": []byte("foo"),
				},
			},
			expected: defaultConnectionDetails,
		},
		{
			description: "nil sslmode",
			args: args{
				credentials: map[string][]byte{
					"sslmode": nil,
				},
			},
			expected: emptyConnectionDetails,
		},
		{
			description: "supported sslmode",
			args: args{
				credentials: map[string][]byte{
					"sslmode": []byte("disable"),
				},
			},
			expected: &ConnectionDetails{
				Port: defaultMSSQLPort,
				SSLParams: map[string]string{
					"encrypt": "disable",
				},
			},
		},
		{
			description: "unsupported sslmode",
			args: args{
				credentials: map[string][]byte{
					"sslmode": []byte("foobar"),
				},
			},
			expected: emptyConnectionDetails,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actualConnDetails := NewConnectionDetails(tc.args.credentials)

			assert.Equal(t, tc.expected, actualConnDetails)
		})
	}
}

func TestDefaultSSLModeExists(t *testing.T) {
	assert.NotEmpty(t, sslModeToBaseParams[string(defaultSSLMode)])
}

func TestConnectionDetails_NewSSLOptions(t *testing.T) {
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
			actualSSLOptions := newSSLParams(tc.args.credentials)

			assert.Equal(t, tc.expected, actualSSLOptions)
		})
	}
}
