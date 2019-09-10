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

	expectedBackendConfig := BackendConfig{
		Host:       "myhost",
		Port:       "1234",
		Username:   "myusername",
		Password:   "mypassword",
		Options:    map[string]string{},
		SSLOptions: map[string]string{},
	}

	actualBackendConfig, err := NewBackendConfig(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedBackendConfig, *actualBackendConfig)
	}
}

func TestSSLOptions(t *testing.T) {
	options := map[string][]byte{
		"host":     []byte("myhost"),
		"port":     []byte("1234"),
		"username": []byte("myusername"),
		"password": []byte("mypassword"),

		"sslrootcert": []byte("mysslrootcert"),
		"sslmode":     []byte("mysslmode"),
		"sslkey":      []byte("mysslkey"),
		"sslcert":     []byte("mysslcert"),
	}

	expectedBackendConfig := BackendConfig{
		Host:     "myhost",
		Port:     "1234",
		Username: "myusername",
		Password: "mypassword",
		Options:  map[string]string{},
		SSLOptions: map[string]string{
			"sslrootcert": "mysslrootcert",
			"sslmode":     "mysslmode",
			"sslkey":      "mysslkey",
			"sslcert":     "mysslcert",
		},
	}

	actualBackendConfig, err := NewBackendConfig(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedBackendConfig, *actualBackendConfig)
	}
}

func TestDefaultPort(t *testing.T) {
	options := map[string][]byte{
		"host":     []byte("myhost"),
		"username": []byte("myusername"),
		"password": []byte("mypassword"),
	}

	expectedBackendConfig := BackendConfig{
		Host:       "myhost",
		Port:       DefaultPostgresPort,
		Username:   "myusername",
		Password:   "mypassword",
		Options:    map[string]string{},
		SSLOptions: map[string]string{},
	}

	actualBackendConfig, err := NewBackendConfig(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedBackendConfig, *actualBackendConfig)
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

	actualBackendConfig, err := NewBackendConfig(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, expectedOptions, (*actualBackendConfig).Options)
	}
}

func TestAddressCanBeUsedInsteadOfHostAndPort(t *testing.T) {
	options := map[string][]byte{
		"address": []byte("myhost2:12345"),
	}

	actualBackendConfig, err := NewBackendConfig(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, "myhost2", (*actualBackendConfig).Host)
		assert.EqualValues(t, "12345", (*actualBackendConfig).Port)
	}
}

func TestAddressWithoutPortCanBeUsedInsteadOfHostAndPort(t *testing.T) {
	options := map[string][]byte{
		"address": []byte("myhost2"),
	}

	actualBackendConfig, err := NewBackendConfig(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, "myhost2", (*actualBackendConfig).Host)
		assert.EqualValues(t, "5432", (*actualBackendConfig).Port)
	}
}

func TestAddress(t *testing.T) {
	options := map[string][]byte{
		"host": []byte("myhost2"),
		"port": []byte("12345"),
	}

	actualBackendConfig, err := NewBackendConfig(options)
	assert.Nil(t, err)

	if err == nil {
		assert.EqualValues(t, "myhost2:12345", (*actualBackendConfig).Address())
	}
}
