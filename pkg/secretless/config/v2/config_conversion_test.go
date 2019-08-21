package v2

//TODO: should we throw custom errors?
import (
	"testing"

	"github.com/stretchr/testify/assert"

	config_v1 "github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
)

func v2DbExample() *Config {
	return &Config{
		Services: []*Service{
			{
				Name:      "test-db",
				Connector: "pg",
				ListenOn:  "tcp://0.0.0.0:2345",
				Credentials: []*Credential{
					{
						Name: "TestSecret1",
						From: "conjur",
						Get:  "some-id-1",
					},
					{
						Name: "TestSecret2",
						From: "literal",
						Get:  "some-id-2",
					},
				},
				ConnectorConfig: nil,
			},
		},
	}
}

func v2HttpExample() *Config {
	return &Config{
		Services: []*Service{
			{
				Name:      "test-http",
				Connector: "aws",
				ListenOn:  "tcp://0.0.0.0:2345",
				Credentials: []*Credential{
					{
						Name: "TestSecret1",
						From: "conjur",
						Get:  "some-id-1",
					},
					{
						Name: "TestSecret2",
						From: "literal",
						Get:  "some-id-2",
					},
				},
				ConnectorConfig: []byte(`
{
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
		v1, err := NewV1ConfigFromV2Config(v2)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		expectedURLs := []string{"^http://aws*", "amzn.com"}
		assert.Equal(t, "aws", v1.Handlers[0].Type)
		assert.ElementsMatch(t, expectedURLs, v1.Handlers[0].Match)
	})

	t.Run("nil config errors", func(t *testing.T) {
		v2 := v2HttpExample()
		v2.Services[0].ConnectorConfig = nil
		_, err := NewV1ConfigFromV2Config(v2)
		assert.Error(t, err)
	})

	t.Run("missing authenticateURLsMatching errors", func(t *testing.T) {
		v2 := v2HttpExample()
		v2.Services[0].ConnectorConfig = []byte(`
{
}`)
		_, err := NewV1ConfigFromV2Config(v2)
		assert.Error(t, err)
	})

	t.Run("all valid auth strategies accepted", func(t *testing.T) {
		v2 := v2HttpExample()

		for _, strategy := range HTTPAuthenticationStrategies {
			v2.Services[0].Connector = strategy.(string)
			_, err := NewV1ConfigFromV2Config(v2)
			assert.NoError(t, err)
		}
	})

	t.Run("authenticateURLsMatching accepts strings", func(t *testing.T) {
		v2 := v2HttpExample()
		v2.Services[0].ConnectorConfig = []byte(`
{
	"authenticateURLsMatching": "^http://aws*"
}`)
		v1, err := NewV1ConfigFromV2Config(v2)

		assert.NoError(t, err)
		if err != nil {
			return
		}

		expectedURLs := []string{"^http://aws*"}
		assert.ElementsMatch(t, expectedURLs, v1.Handlers[0].Match)
	})
}

func TestListenOnConversion(t *testing.T) {

	t.Run("tcp listenOn maps to Address", func(t *testing.T) {
		v2 := v2DbExample()
		v1, err := NewV1ConfigFromV2Config(v2)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, "0.0.0.0:2345", v1.Listeners[0].Address)
	})

	t.Run("unix listenOn maps to Socket", func(t *testing.T) {
		v2 := v2DbExample()
		v2.Services[0].ListenOn = "unix:///some/socket/path"
		v1, err := NewV1ConfigFromV2Config(v2)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.NotNil(t, v1.Listeners[0].Socket)
		assert.Equal(t, "/some/socket/path", v1.Listeners[0].Socket)
	})

	t.Run("unknown listenOn returns error", func(t *testing.T) {
		v2 := v2DbExample()
		v2.Services[0].ListenOn = "/some/socket/path"
		_, err := NewV1ConfigFromV2Config(v2)
		assert.Error(t, err)

		v2.Services[0].ListenOn = "0.0.0.0:2345"
		_, err = NewV1ConfigFromV2Config(v2)
		assert.Error(t, err)
	})
}

func TestCredentialsConversion(t *testing.T) {
	t.Run("Service Credentials map to Handler Credentials", func(t *testing.T) {
		v2cfg := v2DbExample()
		v1cfg, err := NewV1ConfigFromV2Config(v2cfg)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, []config_v1.StoredSecret{
			{
				Name:     "TestSecret1",
				Provider: "conjur",
				ID:       "some-id-1",
			},
			{
				Name:     "TestSecret2",
				Provider: "literal",
				ID:       "some-id-2",
			},
		}, v1cfg.Handlers[0].Credentials)
	})
}
