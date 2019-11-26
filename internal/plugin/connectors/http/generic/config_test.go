package generic

import (
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

		headers, err := cfg.renderedHeaders(map[string][]byte{
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
