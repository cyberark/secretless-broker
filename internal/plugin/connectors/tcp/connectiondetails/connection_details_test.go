package connectiondetails

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type args struct {
	credentials map[string][]byte
}

var defaultTestPort = "1234"

var defaultConnectionDetails = &ConnectionDetails{
	Username: "herp",
	Password: "derp",
	Host:     "0.0.0.0",
	Port:     "1234",
	Options: map[string]string{
		"sslmode":"require",
		"sslrootcert":"foo",
	},
}

var emptyConnectionDetails = &ConnectionDetails{
	Port: defaultTestPort,
	Options: map[string]string{},
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
				Options: tc.fields.Options,
			}

			assert.Equal(t, tc.expected, cd.Address())
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
			description: "nil credentials",
			args: args{
				credentials: map[string][]byte{
				},
			},
			expected: emptyConnectionDetails,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actualConnDetails, _ := NewConnectionDetails(
				tc.args.credentials,
				defaultTestPort,
				)

			assert.Equal(t, tc.expected, actualConnDetails)
		})
	}
}
