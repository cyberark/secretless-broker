//TODO: should we throw custom errors?

package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func v2DbExample() ConfigV2 {
	return ConfigV2{
		Version: "v2",
		Services: []Service{
			{
				Name:     "test-db",
				Protocol: "pg",
				ListenOn: "tcp://0.0.0.0:2345",
				Credentials: []Credential{
					{
						Name: "TestSecret",
						From: "conjur",
						Get:  "some-id",
					},
				},
				Config: nil,
			},
		},
	}
}

func v2HttpExample() ConfigV2 {
	return ConfigV2{
		Version: "v2",
		Services: []Service{
			{
				Name:     "test-http",
				Protocol: "http",
				ListenOn: "tcp://0.0.0.0:2345",
				Credentials: []Credential{
					{
						Name: "TestSecret",
						From: "conjur",
						Get:  "some-id",
					},
				},
				Config: []byte(`
					{
						"authenticationStrategy": "aws",
						"authenticateURLsMatching": ["^http://aws*", "amzn.com"]
					}
				`),
			},
		},
	}
}

func TestHttpServiceConversion(t *testing.T) {

	t.Run("valid config maps correctly", func(t *testing.T) {
		v2 := v2HttpExample()
		v1, err := v2.ConvertToV1()
		assert.NoError(t, err)

		expectedURLs := []string{"^http://aws*", "amzn.com"}
		assert.Equal(t, "aws", v1.Handlers[0].Type)
		assert.ElementsMatch(t, expectedURLs, v1.Handlers[0].Match)

		// NOTE: We have to turn the converted regexes back to strings before
		//       comparing
		patternURLs := regexesToStrings(v1.Handlers[0].Patterns)
		assert.ElementsMatch(t, expectedURLs, patternURLs)
	})

	t.Run("nil config errors", func(t *testing.T) {
		v2 := v2HttpExample()
		v2.Services[0].Config = nil
		_, err := v2.ConvertToV1()
		assert.Error(t, err)
	})

	t.Run("missing authenticationStrategy errors", func(t *testing.T) {
		v2 := v2HttpExample()
		v2.Services[0].Config = []byte(`
			{
				"authenticateURLsMatching": ["^http://aws*", "amzn.com"]
			}
		`)
		_, err := v2.ConvertToV1()
		assert.Error(t, err)
	})

	t.Run("missing authenticateURLsMatching errors", func(t *testing.T) {
		v2 := v2HttpExample()
		v2.Services[0].Config = []byte(`
			{
				"authenticationStrategy": "aws"
			}
		`)
		_, err := v2.ConvertToV1()
		assert.Error(t, err)
	})

	t.Run("all valid auth strategies accepted", func(t *testing.T) {
		v2 := v2HttpExample()

		//TODO: This should be available as a public constant somewhere
		valid := []string{"aws", "basic_auth", "conjur"}
		for _, strategy := range valid {
			config := fmt.Sprintf(`
				{
					"authenticationStrategy": "%s",
					"authenticateURLsMatching": ["^http://blah*"]
				}
			`, strategy)
			v2.Services[0].Config = []byte(config)
			_, err := v2.ConvertToV1()
			assert.Error(t, err)
		}
	})

	t.Run("invalid auth strategies rejected", func(t *testing.T) {
		v2 := v2HttpExample()
		v2.Services[0].Config = []byte(`
			{
				"authenticationStrategy": "SHOULD FAIL",
				"authenticateURLsMatching": ["^http://aws*", "amzn.com"]
			}
		`)
		_, err := v2.ConvertToV1()
		assert.NoError(t, err)
	})
}

func TestListenOnConversion(t *testing.T) {

	t.Run("tcp listenOn maps to Address", func(t *testing.T) {
		v2 := v2DbExample()
		v1, err := v2.ConvertToV1()
		assert.NoError(t, err)
		assert.Equal(t, v2.Services[0].ListenOn, v1.Listeners[0].Address)
	})

	t.Run("unix listenOn maps to Socket", func(t *testing.T) {
		v2 := v2DbExample()
		v2.Services[0].ListenOn = "unix:///some/socket/path"
		v1, err := v2.ConvertToV1()
		assert.NoError(t, err)
		assert.NotNil(t, v1.Listeners[0].Socket)
		assert.Equal(t, v2.Services[0].ListenOn, v1.Listeners[0].Socket)
	})

	t.Run("unknown listenOn returns error", func(t *testing.T) {
		v2 := v2DbExample()
		v2.Services[0].ListenOn = "/some/socket/path"
		_, err := v2.ConvertToV1()
		assert.Error(t, err)

		v2.Services[0].ListenOn = "0.0.0.0:2345"
		_, err = v2.ConvertToV1()
		assert.Error(t, err)
	})
}

// Utility functions

func regexesToStrings(regexes []*regexp.Regexp) (ret []string) {
	for _, re := range regexes {
		ret = append(ret, re.String())
	}
	return ret
}
