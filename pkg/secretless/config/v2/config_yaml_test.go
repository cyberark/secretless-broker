package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const validConfigYAML = `
services:
  pg_tcp:
    protocol: pg
    connector: pg
    listenOn: tcp://0.0.0.0:15432
    credentials:
      username: test
      password:
        from: env
        get: PG_PASSWORD
    config:
      a: b
`

func Test_newConfigYAML(t *testing.T) {
	t.Run("errors on empty YAML byte slice", func(t *testing.T) {
		_, err := newConfigYAML([]byte(
			`
`))
		assert.Error(t, err)
	})

	t.Run("errors on invalid YAML byte slice", func(t *testing.T) {
		_, err := newConfigYAML([]byte(
			`
services:
	serviceName: not service dictionary
`))
		assert.Error(t, err)
	})

	t.Run("does not error on valid YAML byte slice", func(t *testing.T) {
		_, err := newConfigYAML([]byte(validConfigYAML))
		assert.NoError(t, err)
	})

	t.Run("creates configYAML struct on valid YAML byte slice", func(t *testing.T) {
		cfg, _ := newConfigYAML([]byte(validConfigYAML))
		assert.Equal(t, &configYAML{
			Services: map[string]*serviceYAML{
				"pg_tcp": {
					Protocol:  "pg",
					Connector: "pg",
					ListenOn:  "tcp://0.0.0.0:15432",
					Credentials: credentialsYAML{
						"username": "test",
						"password": map[interface{}]interface{}{
							"from": "env",
							"get":  "PG_PASSWORD",
						},
					},
					Config: map[interface{}]interface{}{
						"a": "b",
					},
				},
			},
		}, cfg)
	})
}
