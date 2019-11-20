package generic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func sampleConfig() []byte {
	return []byte(`
  credentialValidation:
    username: '[^:]+'
  headers:
    Authorization: 'Basic {{ printf "%s:%s" .username .password | base64 }}'
    Name-With-Dashes: '{{ .username }}'
    SimpleConcatenation: '{{ .username }} - {{ .password }}'
  forceSSL: true
  authenticateURLsMatching:
    - ^http
`)
}

func Test_newConfig(t *testing.T) {
	t.Run("creates expected headers", func(t *testing.T) {
		cfgYAML, err := NewConfigYAML(sampleConfig())
		if err != nil {
			assert.Fail(t, "sampleConfig should never fail")
			return
		}
		cfg, err := newConfig(cfgYAML)

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
// TODO: Add validation tests
