package generic

import (
	"net/url"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func sampleConfigYAML() []byte {
	return []byte(`
  credentialValidations:
    username: '^[^:]+$'
  headers:
    Authorization: 'Basic {{ printf "%s:%s" .username .password | base64 }}'
    Name-With-Dashes: '{{ .username }}'
    SimpleConcatenation: '{{ .username }} - {{ .password }}'
  queryParams:
    Key: '{{ .key | base64 }}'
    NameWithSpaces: '{{ .name }}'
  forceSSL: true
  authenticateURLsMatching:
    - ^http
`)
}

func sampleConfig() (*config, error) {
	cfgYAML, err := NewConfigYAML(sampleConfigYAML())
	if err != nil {
		return nil, errors.Wrap(err, "error parsing sample config YAML")
	}
	cfg, err := newConfig(cfgYAML)
	if err != nil {
		return nil, errors.Wrap(err, "error calling newConfig()")
	}
	return cfg, err
}

func Test_newConfig(t *testing.T) {
	t.Run("creates expected headers", func(t *testing.T) {
		cfg, err := sampleConfig()
		assert.NoError(t, err)
		if err != nil {
			return
		}

		headers, err := renderTemplates(
			cfg.Headers,
			map[string][]byte{
				"username": []byte("Jonah"),
				"password": []byte("secret"),
			})

		assert.NoError(t, err)
		if err != nil {
			return
		}

		// A couple simple test cases
		assert.Equal(t, "Jonah", headers["Name-With-Dashes"])
		assert.Equal(t, "Jonah - secret", headers["SimpleConcatenation"])

		// Assert against value calculated with independent base64 encoder.
		assert.Equal(t, "Basic Sm9uYWg6c2VjcmV0", headers["Authorization"])
	})

	t.Run("creates expected query params", func(t *testing.T) {
		cfg, err := sampleConfig()
		assert.NoError(t, err)
		if err != nil {
			return
		}

		params, err := renderTemplates(
			cfg.QueryParams,
			map[string][]byte{
				"key":  []byte("someKey"),
				"name": []byte("Foo Bar"),
			})

		assert.NoError(t, err)
		if err != nil {
			return
		}

		// A simple test case
		assert.Equal(t, "Foo Bar", params["NameWithSpaces"])

		// Assert against value calculated with independent base64 encoder.
		assert.Equal(t, "c29tZUtleQ==", params["Key"])
	})
}

func Test_validate(t *testing.T) {
	testCases := []struct {
		description string
		creds       map[string][]byte
		expErrStr   string
	}{
		{
			description: "validates good credentials",
			creds: map[string][]byte{
				"username": []byte("Jonah"),
				"password": []byte("secret"),
			},
			expErrStr: "",
		},
		{
			description: "errors on missing required credential",
			creds: map[string][]byte{
				"password": []byte("secret"),
			},
			expErrStr: "missing required credential",
		},
		{
			description: "errors on invalid credential",
			creds: map[string][]byte{
				"username": []byte("Jon:ah"),
				"password": []byte("secret"),
			},
			expErrStr: "doesn't match pattern",
		},
	}

	cfg, err := sampleConfig()
	assert.NoError(t, err)
	if err != nil {
		return
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err = cfg.validate(tc.creds)
			if tc.expErrStr == "" {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			if err != nil {
				assert.True(t, strings.Contains(err.Error(), tc.expErrStr))
			}
		})
	}
}

func Test_appendQueryParams(t *testing.T) {
	testCases := []struct {
		description      string
		params           map[string]string
		url              url.URL
		expectedRawQuery string
	}{
		{
			description: "appends params without replacing existing params",
			params: map[string]string{
				"foo": "bar",
			},
			url: url.URL{
				RawQuery: "some=place",
			},
			expectedRawQuery: "foo=bar&some=place",
		},
		{
			description: "special characters encode properly",
			params: map[string]string{
				"space":   "bar biz",
				"key":     "abc==",
				"special": "@#$%^&*(){}[].,+",
			},
			url: url.URL{
				RawQuery: "some=place",
			},
			expectedRawQuery: "key=abc%3D%3D&some=place&space=bar+biz&special=%40%23%24" +
				"%25%5E%26%2A%28%29%7B%7D%5B%5D.%2C%2B",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actualRawQuery := appendQueryParams(tc.url, tc.params)

			assert.Equal(t, tc.expectedRawQuery, actualRawQuery)
		})
	}
}
