package config

import (
	"fmt"
	"regexp"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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

	Convey("Compiles the handler match expressions into patterns", t, func() {
		yaml := `
listeners:
  - name: conjur
    protocol: http
    address: 0.0.0.0:1080

handlers:
  - name: conjur
    match: [ ".*" ]
`
		pattern, err := regexp.Compile(".*")
		So(err, ShouldBeNil)

		config, err := Load([]byte(yaml))
		So(err, ShouldBeNil)
		So(config.Handlers[0].Patterns[0].String(), ShouldEqual, pattern.String())
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
		So(config.Handlers, ShouldHaveLength, 1)
		So(config.Listeners, ShouldHaveLength, 1)
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
		So(config.Handlers, ShouldHaveLength, 1)
		So(config.Listeners, ShouldHaveLength, 1)
	})

	Convey("Reports an invalid top-level map key", t, func() {
		yaml := `
foobar: []
`
		_, err := Load([]byte(yaml))
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "field foobar not found in struct config.Config")
	})

	Convey("Reports an unnamed Listener definition", t, func() {
		yaml := `
listeners:
  - protocol: pg
`
		_, err := Load([]byte(yaml))
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Listeners: (0: (Name: cannot be blank.).)")
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
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Listeners: (0: must have an Address or Socket.)")
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
		So(fmt.Sprintf("%s", err), ShouldContainSubstring, "Handlers: (0: (Name: cannot be blank.).)")
	})
}
