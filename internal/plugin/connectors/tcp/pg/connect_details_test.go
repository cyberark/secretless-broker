package pg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpectedFields(t *testing.T) {
	options := map[string][]byte{
		"host":     []byte("myhost"),
		"port":     []byte("1234"),
		"username": []byte("myusername"),
		"password": []byte("mypassword"),
	}

	expectedConnectionDetails := ConnectionDetails{
		Host:     "myhost",
		Port:     "1234",
		Username: "myusername",
		Password: "mypassword",
		Options:  map[string]string{},
		SSLOptions: map[string]string{
			"host": "myhost",
		},
	}

	actualConnectionDetails, err := NewConnectionDetails(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedConnectionDetails, *actualConnectionDetails)
	}
}

func TestSSLOptions(t *testing.T) {
	options := map[string][]byte{
		"host":     []byte("myhost"),
		"port":     []byte("1234"),
		"username": []byte("myusername"),
		"password": []byte("mypassword"),

		"sslhost":     []byte("customhost"),
		"sslrootcert": []byte("mysslrootcert"),
		"sslmode":     []byte("mysslmode"),
		"sslkey":      []byte("mysslkey"),
		"sslcert":     []byte("mysslcert"),
	}

	expectedConnectionDetails := ConnectionDetails{
		Host:     "myhost",
		Port:     "1234",
		Username: "myusername",
		Password: "mypassword",
		Options:  map[string]string{},
		SSLOptions: map[string]string{
			"host":        "myhost",
			"sslhost":     "customhost",
			"sslrootcert": "mysslrootcert",
			"sslmode":     "mysslmode",
			"sslkey":      "mysslkey",
			"sslcert":     "mysslcert",
		},
	}

	actualConnectionDetails, err := NewConnectionDetails(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedConnectionDetails, *actualConnectionDetails)
	}
}

func TestDefaultPort(t *testing.T) {
	options := map[string][]byte{
		"host":     []byte("myhost"),
		"username": []byte("myusername"),
		"password": []byte("mypassword"),
	}

	expectedConnectionDetails := ConnectionDetails{
		Host:     "myhost",
		Port:     DefaultPostgresPort,
		Username: "myusername",
		Password: "mypassword",
		Options:  map[string]string{},
		SSLOptions: map[string]string{
			"host": "myhost",
		},
	}

	actualConnectionDetails, err := NewConnectionDetails(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedConnectionDetails, *actualConnectionDetails)
	}
}

func TestUnexpectedFieldsAreSavedAsOptions(t *testing.T) {
	options := map[string][]byte{
		"host":     []byte("myhost"),
		"port":     []byte("1234"),
		"foo":      []byte("3306"),
		"username": []byte("myusername"),
		"bar":      []byte("data"),
		"password": []byte("mypassword"),
	}

	expectedOptions := map[string]string{
		"foo": "3306",
		"bar": "data",
	}

	actualConnectionDetails, err := NewConnectionDetails(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedOptions, (*actualConnectionDetails).Options)
	}
}

func TestAddressCanBeUsedInsteadOfHostAndPort(t *testing.T) {
	options := map[string][]byte{
		"address": []byte("myhost2:12345"),
	}

	actualConnectionDetails, err := NewConnectionDetails(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, "myhost2", (*actualConnectionDetails).Host)
		assert.EqualValues(t, "12345", (*actualConnectionDetails).Port)
	}
}

func TestAddressWithoutPortCanBeUsedInsteadOfHostAndPort(t *testing.T) {
	options := map[string][]byte{
		"address": []byte("myhost2"),
	}

	actualConnectionDetails, err := NewConnectionDetails(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, "myhost2", (*actualConnectionDetails).Host)
		assert.EqualValues(t, "5432", (*actualConnectionDetails).Port)
	}
}

func TestAddress(t *testing.T) {
	options := map[string][]byte{
		"host": []byte("myhost2"),
		"port": []byte("12345"),
	}

	actualConnectionDetails, err := NewConnectionDetails(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, "myhost2:12345", (*actualConnectionDetails).Address())
	}
}
