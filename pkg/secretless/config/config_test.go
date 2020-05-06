package config

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	crd_api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
)

func Test_Config(t *testing.T) {
	Convey("Reports absence of handlers", t, func() {
		yaml := `
---
`
		_, err := Load([]byte(yaml))
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Handlers: cannot be blank")
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Listeners: cannot be blank")
	})

	Convey("Loads a realistic configuration without errors", t, func() {
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
		So(err, ShouldBeNil)
		So(config.Services, ShouldHaveLength, 1)
	})

	Convey("Allows listeners to have debug flag", t, func() {
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
		So(err, ShouldBeNil)
		So(config.Services, ShouldHaveLength, 1)
	})

	Convey("Reports an unnamed Listener definition", t, func() {
		yaml := `
listeners:
  - protocol: pg
`
		_, err := Load([]byte(yaml))
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Name: cannot be blank")
	})

	Convey("Reports an unknown protocol", t, func() {
		yaml := `
listeners:
  - protocol: myapp
`
		_, err := Load([]byte(yaml))
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Name: cannot be blank.")
	})

	Convey("Reports a Handler which wants to use an undefined Listener", t, func() {
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
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Handlers: (0: has no associated listener.)")
	})

	Convey("Reports a Listener without an address or socket", t, func() {
		yaml := `
listeners:
  - name: mylistener
    protocol: pg

handlers:
  - name: mylistener
`
		_, err := Load([]byte(yaml))
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "address or socket is required")
	})

	Convey("Reports an unnamed Handler definition", t, func() {
		yaml := `
listeners:
  - name: http_default
    protocol: tcp

handlers:
  - listener: http_default
`
		_, err := Load([]byte(yaml))
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Name: cannot be blank")
	})

	Convey("Can serialize match fields", t, func() {
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
		So(err, ShouldBeNil)
		So(config.String(), ShouldContainSubstring, "test_for_secretless_issues_216")
	})

	Convey("Can generate config from CRD configuration", t, func() {
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
		So(err, ShouldBeNil)

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
		So(err, ShouldBeNil)
		So(config.String(), ShouldEqual, expectedConfig.String())
	})
}
