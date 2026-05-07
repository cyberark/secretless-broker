package v2

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPConfig(t *testing.T) {
	t.Run("input 'authenticateURLsMatching' converts to regex", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: ".*"
`)
		cfg, _ := NewHTTPConfig(configFileContents)
		assert.Equal(t, cfg.AuthenticateURLsMatching[0], regexp.MustCompile(".*"))
	})

	t.Run("upstreamURL parsed correctly", func(t *testing.T) {
		configFileContents := []byte(`
authenticateURLsMatching: ".*"
upstreamURL: "https://api.stripe.com"
`)
		cfg, err := NewHTTPConfig(configFileContents)
		require.NoError(t, err)
		require.NotNil(t, cfg.UpstreamURL)
		assert.Equal(t, "https", cfg.UpstreamURL.Scheme)
		assert.Equal(t, "api.stripe.com", cfg.UpstreamURL.Host)
	})

	t.Run("upstreamURL omitted is nil", func(t *testing.T) {
		configFileContents := []byte(`
authenticateURLsMatching: ".*"
`)
		cfg, err := NewHTTPConfig(configFileContents)
		require.NoError(t, err)
		assert.Nil(t, cfg.UpstreamURL)
	})

	t.Run("upstreamURL invalid URL returns error", func(t *testing.T) {
		configFileContents := []byte(`
authenticateURLsMatching: ".*"
upstreamURL: "not-a-url"
`)
		_, err := NewHTTPConfig(configFileContents)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "upstreamURL must include scheme and host")
	})

	t.Run("upstreamURL with port", func(t *testing.T) {
		configFileContents := []byte(`
authenticateURLsMatching: ".*"
upstreamURL: "https://api.example.com:8443"
`)
		cfg, err := NewHTTPConfig(configFileContents)
		require.NoError(t, err)
		require.NotNil(t, cfg.UpstreamURL)
		assert.Equal(t, "api.example.com:8443", cfg.UpstreamURL.Host)
	})
}

func TestNewHTTPConfigYAML(t *testing.T) {
	t.Run("'authenticateUrlsMatching' accepts a single string", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: ".*"
`)
		cfg, _ := newHTTPConfigYAML(configFileContents)
		assert.Equal(t, cfg.AuthenticateURLsMatching, []string{".*"})
	})

	t.Run("'authenticateUrlsMatching' accepts an array of strings", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: 
 - "*"
`)
		cfg, _ := newHTTPConfigYAML(configFileContents)
		assert.Equal(t, cfg.AuthenticateURLsMatching, []string{"*"})
	})

	t.Run("error on bad type for 'authenticateURLsMatching' list", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: 
 - true
 - "meow"
`)
		_, err := newHTTPConfigYAML(configFileContents)
		assert.Error(t, err)
	})

	t.Run("error on bad type for 'authenticateURLsMatching' scalar", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: false
`)
		_, err := newHTTPConfigYAML(configFileContents)
		assert.Error(t, err)
	})

	t.Run("error on invalid file contents", func(t *testing.T) {
		configFileContents := []byte(
			`
{
"x": false
}
`)
		_, err := newHTTPConfigYAML(configFileContents)
		assert.Error(t, err)
	})

	t.Run("error on blank file contents", func(t *testing.T) {
		configFileContents := []byte("")
		_, err := NewConfig(configFileContents)
		assert.Error(t, err)
	})

}
