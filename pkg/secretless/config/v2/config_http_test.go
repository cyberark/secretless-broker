package v2

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
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
