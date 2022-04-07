package generic

import (
	"bytes"
	"net/http"
	"testing"

	mockLogger "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func sampleOauth1ConfigYAML() []byte {
	return []byte(`
  oauth1:
    consumer_key: "{{ .consumer_key }}"
    consumer_secret: "{{ .consumer_secret }}"
    token: "{{ .token }}"
    token_secret: "{{ .token_secret }}"
  forceSSL: true
  authenticateURLsMatching:
    - ^http
`)
}

func sampleOauth1Config() (*config, error) {
	cfgYAML, err := NewConfigYAML(sampleOauth1ConfigYAML())
	if err != nil {
		return nil, errors.Wrap(err, "error parsing sample config YAML")
	}
	cfg, err := newConfig(cfgYAML)
	if err != nil {
		return nil, errors.Wrap(err, "error calling newConfig()")
	}
	return cfg, err
}

func plainRequest() *http.Request {
	req, err := http.NewRequest("GET", "http://example.com", &bytes.Buffer{})
	if err != nil {
		panic(err)
	}
	return req
}

func requestWithAuthHeader() *http.Request {
	req := plainRequest()
	req.Header.Set("Authorization", "Basic 1234567890")
	return req
}

func plainConfig() *config {
	cfg, err := sampleConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

func oauth1Config() *config {
	cfg, err := sampleOauth1Config()
	if err != nil {
		panic(err)
	}
	return cfg
}

var testCases = []struct {
	description string
	cfg         *config
	request     *http.Request
	creds       connector.CredentialValuesByID
	expErrStr   string
}{
	{
		description: "happy path",
		cfg:         plainConfig(),
		request:     plainRequest(),
		creds: connector.CredentialValuesByID{
			"username": []byte("someuser"),
			"key":      []byte("someKey"),
			"name":     []byte("Foo Bar"),
		},
	},
	{
		description: "missing username",
		cfg:         plainConfig(),
		request:     plainRequest(),
		creds: connector.CredentialValuesByID{
			"key":  []byte("someKey"),
			"name": []byte("Foo Bar"),
		},
		expErrStr: "missing required credential: \"username\"",
	},
	{
		description: "missing query params",
		cfg:         plainConfig(),
		request:     plainRequest(),
		creds: connector.CredentialValuesByID{
			"username": []byte("someuser"),
		},
		expErrStr: "failed to render query params: Key: couldn't render template",
	},
	{
		description: "oauth1 happy path",
		cfg:         oauth1Config(),
		request:     plainRequest(),
		creds: connector.CredentialValuesByID{
			"consumer_key":    []byte("someConsumerKey"),
			"consumer_secret": []byte("someConsumerSecret"),
			"token":           []byte("someToken"),
			"token_secret":    []byte("someTokenSecret"),
		},
	},
	{
		description: "oauth1 missing params",
		cfg:         oauth1Config(),
		request:     plainRequest(),
		creds: connector.CredentialValuesByID{
			"consumer_key":    []byte(nil),
			"consumer_secret": []byte("someConsumerSecret"),
			"token":           []byte("someToken"),
			"token_secret":    []byte("someTokenSecret"),
		},
		expErrStr: "failed to create oAuth1 'Authorization' header: required oAuth1 parameter 'consumer_key' not found",
	},
	{
		description: "oauth1 duplicate auth header",
		cfg:         oauth1Config(),
		request:     requestWithAuthHeader(),
		creds: connector.CredentialValuesByID{
			"consumer_key":    []byte("someConsumerKey"),
			"consumer_secret": []byte("someConsumerSecret"),
			"token":           []byte("someToken"),
			"token_secret":    []byte("someTokenSecret"),
		},
		expErrStr: "authorization header already exists, cannot override header",
	},
}

func TestConnect(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			logger := mockLogger.NewLogger()
			c := Connector{
				logger: logger,
				config: tc.cfg,
			}

			err := c.Connect(tc.request, tc.creds)

			if tc.expErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErrStr)
				return
			}

			assert.NoError(t, err)
			// Ensure forceSSL is working
			assert.Equal(t, tc.request.URL.Scheme, "https")
		})
	}
}
