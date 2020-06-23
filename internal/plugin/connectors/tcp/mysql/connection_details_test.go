package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpectedFields(t *testing.T) {
	credentials := map[string][]byte{
		"host":     []byte("myhost"),
		"port":     []byte("1234"),
		"username": []byte("myusername"),
		"password": []byte("mypassword"),
		"sslmode":  []byte("disable"),
	}

	expectedConnDetails := ConnectionDetails{
		Host:    "myhost",
		Options: map[string]string{},
		SSLOptions: map[string]string{
			"host":    "myhost",
			"sslmode": "disable",
		},
		Password: "mypassword",
		Port:     1234,
		Username: "myusername",
	}

	actualConnDetails, err := NewConnectionDetails(credentials)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedConnDetails, *actualConnDetails)
	}
}

func TestDefaultPort(t *testing.T) {
	credentials := map[string][]byte{
		"host":     []byte("myhost"),
		"username": []byte("myusername"),
		"password": []byte("mypassword"),
	}

	expectedConnDetails := ConnectionDetails{
		Host:     "myhost",
		Port:     DefaultMySQLPort,
		Username: "myusername",
		Password: "mypassword",
		Options:  map[string]string{},
		SSLOptions: map[string]string{
			"host": "myhost",
		},
	}

	actualConnDetails, err := NewConnectionDetails(credentials)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedConnDetails, *actualConnDetails)
	}
}

func TestUnexpectedFieldsAreSavedAsOptions(t *testing.T) {
	credentials := map[string][]byte{
		"host":     []byte("myhost"),
		"port":     []byte("1234"),
		"foo":      []byte("5432"),
		"username": []byte("myusername"),
		"bar":      []byte("data"),
		"password": []byte("mypassword"),
	}

	expectedOptions := map[string]string{
		"foo": "5432",
		"bar": "data",
	}

	actualConnDetails, err := NewConnectionDetails(credentials)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedOptions, (*actualConnDetails).Options)
	}
}

func TestAddress(t *testing.T) {
	credentials := map[string][]byte{
		"host": []byte("myhost2"),
		"port": []byte("12345"),
	}

	actualConnDetails, err := NewConnectionDetails(credentials)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, "myhost2:12345", (*actualConnDetails).Address())
	}
}
