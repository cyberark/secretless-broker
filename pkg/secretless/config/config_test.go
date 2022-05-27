package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	crd_api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
)

func Test_Config(t *testing.T) {
	t.Run("Reports absence of handlers", func(t *testing.T) {
		yaml := `
---
`
		_, err := Load([]byte(yaml))
		assert.Contains(t, fmt.Sprintf("%s", err), "Handlers: cannot be blank")
		assert.Contains(t, fmt.Sprintf("%s", err), "Listeners: cannot be blank")
	})

	t.Run("Loads a realistic configuration without errors", func(t *testing.T) {
		yaml := `
listeners:
- name: http_default
  protocol: http
  address: 0.0.0.0:1080

handlers:
- name: conjur
  listener: http_default
  credentials:
    - name: accessToken
      provider: conjur
      id: accessToken

`
		config, err := Load([]byte(yaml))
		assert.NoError(t, err)
		assert.Len(t, config.Services, 1)
	})

	t.Run("Allows listeners to have debug flag", func(t *testing.T) {
		yaml := `
listeners:
- name: http_default
  protocol: http
  debug: true
  address: 0.0.0.0:1080

handlers:
- name: conjur
  listener: http_default
  credentials:
    - name: accessToken
      provider: conjur
      id: accessToken

`
		config, err := Load([]byte(yaml))
		assert.NoError(t, err)
		assert.Len(t, config.Services, 1)
	})

	t.Run("Reports an unnamed Listener definition", func(t *testing.T) {
		yaml := `
listeners:
  - protocol: pg
`
		_, err := Load([]byte(yaml))
		assert.Contains(t, fmt.Sprintf("%s", err), "Name: cannot be blank")
	})

	t.Run("Reports an unknown protocol", func(t *testing.T) {
		yaml := `
listeners:
  - protocol: myapp
`
		_, err := Load([]byte(yaml))
		assert.Contains(t, fmt.Sprintf("%s", err), "Name: cannot be blank")
	})

	t.Run("Reports a Handler which wants to use an undefined Listener", func(t *testing.T) {
		yaml := `
listeners:
  - name: http_default
    protocol: http
    address: 0.0.0.0:1080

handlers:
  - name: myhandler
    listener: none
`
		_, err := Load([]byte(yaml))
		assert.Contains(t, fmt.Sprintf("%s", err), "Handlers: (0: has no associated listener.)")
	})

	t.Run("Reports a Listener without an address or socket", func(t *testing.T) {
		yaml := `
listeners:
  - name: mylistener
    protocol: pg

handlers:
  - name: mylistener
`
		_, err := Load([]byte(yaml))
		assert.Contains(t, fmt.Sprintf("%s", err), "address or socket is required")
	})

	t.Run("Reports an unnamed Handler definition", func(t *testing.T) {
		yaml := `
listeners:
  - name: http_default
    protocol: tcp

handlers:
  - listener: http_default
`
		_, err := Load([]byte(yaml))
		assert.Contains(t, fmt.Sprintf("%s", err), "Name: cannot be blank")
	})

	t.Run("Can serialize match fields", func(t *testing.T) {
		yaml := `
listeners:
  - name: http_default
    protocol: http
    address: 0.0.0.0:1080

handlers:
  - name: http_default
    listener: http_default
    credentials:
      - name: accessToken
        provider: conjur
        id: accessToken
    match:
      - test_for_secretless_issues_216
`
		config, err := Load([]byte(yaml))
		assert.NoError(t, err)
		assert.Contains(t, config.String(), "test_for_secretless_issues_216")
	})

	t.Run("Can generate config from CRD configuration", func(t *testing.T) {
		expectedConfigYaml := `
listeners:
  - name: http_default
    protocol: http
    address: 0.0.0.0:1080

handlers:
  - name: http_default_handler
    listener: http_default
    credentials:
    - name: accessToken
      provider: conjur
      id: accessToken
    match:
    - http://*
`

		// We implicitly rely on Load to work properly for this test to pass
		expectedConfig, err := Load([]byte(expectedConfigYaml))
		assert.NoError(t, err)

		// Create an API object that would be similar to one used to trigger a config reload
		crdConfig := crd_api_v1.Configuration{
			Spec: crd_api_v1.ConfigurationSpec{
				Handlers: []crd_api_v1.Handler{
					crd_api_v1.Handler{
						Name:         "http_default_handler",
						ListenerName: "http_default",
						Match: []string{
							"http://*",
						},
						Credentials: []crd_api_v1.Variable{
							{
								Name:     "accessToken",
								Provider: "conjur",
								ID:       "accessToken",
							},
						},
					},
				},
				Listeners: []crd_api_v1.Listener{
					crd_api_v1.Listener{
						Name:     "http_default",
						Protocol: "http",
						Address:  "0.0.0.0:1080",
					},
				},
			},
		}
		config, err := LoadFromCRD(crdConfig)
		assert.NoError(t, err)
		assert.Equal(t, expectedConfig.String(), config.String())
	})
}
