package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPConfig(t *testing.T) {
	t.Run("http config hydration with 'authenticateURLsMatching' string", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: "*"
`)
		cfg, _ := newHTTPConfig(configFileContents)
		assert.Equal(t, cfg.AuthenticateURLsMatching, []string{"*"})
	})

	t.Run("http config hydration with 'authenticateURLsMatching' string list", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: 
 - "*"
`)
		cfg, _ := newHTTPConfig(configFileContents)
		assert.Equal(t, cfg.AuthenticateURLsMatching, []string{"*"})
	})

	t.Run("error on bad type for 'authenticateURLsMatching' list", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: 
 - true
 - "meow"
`)
		_, err := newHTTPConfig(configFileContents)
		assert.Error(t, err)
	})

	t.Run("error on bad type for 'authenticateURLsMatching' scalar", func(t *testing.T) {
		configFileContents := []byte(
			`
authenticateURLsMatching: false
`)
		_, err := newHTTPConfig(configFileContents)
		assert.Error(t, err)
	})

	t.Run("error on invalid file contents", func(t *testing.T) {
		configFileContents := []byte(
`
{
"x": false
}
`)
		_, err := newHTTPConfig(configFileContents)
		assert.Error(t, err)
	})

	t.Run("error on blank file contents", func(t *testing.T) {
		configFileContents := []byte("")
		_, err := NewConfig(configFileContents)
		assert.Error(t, err)
	})

}
