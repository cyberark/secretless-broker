package mssql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnectionDetails_Address(t *testing.T) {
	type fields struct {
		Host     string
		Port     uint
		Username string
		Password string
		SSLMode  string
	}
	tests := []struct {
		description string
		fields      fields
		expected    string
	}{
		{
			description: "default address format",
			fields: fields{
				Host:     "hostname",
				Port:     5555,
				Username: "foo",
				Password: "bar",
				SSLMode:  "xyz",
			},
			expected: "hostname:5555",
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
	type args struct {
		credentials map[string][]byte
	}
	tests := []struct {
		description string
		args        args
		expected    *ConnectionDetails
	}{
		{
			description: "ssl mode is empty - use default",
			args: args{
				credentials: map[string][]byte{
					"sslmode": nil,
				},
			},
			expected: &ConnectionDetails{
				Host:     "",
				Port:     defaultMSSQLPort,
				Username: "",
				Password: "",
				SSLMode:  "disable",
			},
		},
		{
			description: "ssl mode is disable - use value",
			args: args{
				credentials: map[string][]byte{
					"sslmode": []byte("enable"),
				},
			},
			expected: &ConnectionDetails{
				Host:     "",
				Port:     defaultMSSQLPort,
				Username: "",
				Password: "",
				SSLMode:  "disable",
			},
		},
		{
			description: "ssl mode is unsupported - use default",
			args: args{
				credentials: map[string][]byte{
					"sslmode": []byte("foobar"),
				},
			},
			expected: &ConnectionDetails{
				Host:     "",
				Port:     defaultMSSQLPort,
				Username: "",
				Password: "",
				SSLMode:  "disable",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actualConnDetails := NewConnectionDetails(tt.args.credentials)

			assert.Equal(t, tt.expected, actualConnDetails)

			// Verify that credentials have been zeroed
			for cred := range tt.args.credentials {
				assert.Empty(t, cred)
			}
		})
	}
}
