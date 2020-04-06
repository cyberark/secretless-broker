package mssql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type args struct {
	credentials map[string][]byte
}

var defaultConnectionDetails = &ConnectionDetails{
	Username: "herp",
	Password: "derp",
	Host:     "0.0.0.0",
	Port:     1234,
	SSLMode:  "disable",
	SSLOptions: map[string]string{
		"sslrootcert": "foo",
		"sslkey":      "bar",
		"sslcert":     "foobar",
	},
}

var emptyConnectionDetails = &ConnectionDetails{
	Port:       defaultMSSQLPort,
	SSLMode:    "disable",
	SSLOptions: map[string]string{},
}

func TestConnectionDetails_Address(t *testing.T) {
	tests := []struct {
		description string
		fields      *ConnectionDetails
		expected    string
	}{
		{
			description: "default address format",
			fields:      defaultConnectionDetails,
			expected:    "0.0.0.0:1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			cd := &ConnectionDetails{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
				SSLMode:  tt.fields.SSLMode,
			}

			assert.Equal(t, tt.expected, cd.Address())
		})
	}
}

func TestNewConnectionDetails(t *testing.T) {
	tests := []struct {
		description string
		args        args
		expected    *ConnectionDetails
	}{
		{
			description: "standard case - all values filled",
			args: args{
				credentials: map[string][]byte{
					"username":    []byte("herp"),
					"password":    []byte("derp"),
					"host":        []byte("0.0.0.0"),
					"port":        []byte("1234"),
					"sslmode":     []byte("require"),
					"sslrootcert": []byte("foo"),
					"sslkey":      []byte("bar"),
					"sslcert":     []byte("foobar"),
				},
			},
			expected: defaultConnectionDetails,
		},
		{
			description: "ssl mode is empty - use default",
			args: args{
				credentials: map[string][]byte{
					"sslmode": nil,
				},
			},
			expected: emptyConnectionDetails,
		},
		{
			description: "ssl mode is disable - use value",
			args: args{
				credentials: map[string][]byte{
					"sslmode": []byte("enable"),
				},
			},
			expected: emptyConnectionDetails,
		},
		{
			description: "ssl mode is unsupported - use default",
			args: args{
				credentials: map[string][]byte{
					"sslmode": []byte("foobar"),
				},
			},
			expected: emptyConnectionDetails,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actualConnDetails := NewConnectionDetails(tt.args.credentials)

			assert.Equal(t, tt.expected, actualConnDetails)
		})
	}
}

func TestConnectionDetails_NewSSLOptions(t *testing.T) {
	tests := []struct {
		description string
		args        args
		expected    map[string]string
	}{
		{
			description: "no values",
			args: args{
				credentials: map[string][]byte{},
			},
			expected: map[string]string{},
		},
		{
			description: "standard values found",
			args: args{
				credentials: map[string][]byte{
					"sslrootcert": []byte("foo"),
					"sslkey":      []byte("bar"),
					"sslcert":     []byte("foobar"),
				},
			},
			expected: map[string]string{
				"sslrootcert": "foo",
				"sslkey":      "bar",
				"sslcert":     "foobar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actualSSLOptions := newSSLOptions(tt.args.credentials)

			assert.Equal(t, tt.expected, actualSSLOptions)
		})
	}
}
