package v2

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestConfig_MarshalYAML(t *testing.T) {
	out, err := yaml.Marshal(&Config{
		Services: []*Service{{
			Connector:       "connector",
			ConnectorConfig: []byte(`{"a": 1, "b": [1,2,3], "c": "xyz"}`),
			Credentials: []*Credential{{
				Name: "name",
				From: "from",
				Get:  "get",
			}},
			ListenOn: "tcp://0.0.0.0:8080",
			Name:     "name",
		}},
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, string(out), strings.TrimSpace(`
version: "2"
services:
  name:
    protocol: ""
    connector: connector
    listenOn: tcp://0.0.0.0:8080
    credentials:
      name:
        from: from
        get: get
    config:
      a: 1
      b:
      - 1
      - 2
      - 3
      c: xyz
`)+"\n")
}
