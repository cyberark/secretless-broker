package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi/response"
	"github.com/stretchr/testify/assert"

	_ "github.com/joho/godotenv/autoload"
)

// TestConjur_Handler verifies that Conjur API requests which are proxied through the Secretless
// handler do not require authentication credentials.
func TestConjur_Handler(t *testing.T) {
	t.Run("Can fetch a variable value", func(t *testing.T) {
		variableURL := fmt.Sprintf("%s/secrets/%s/variable/db/password", os.Getenv("CONJUR_APPLIANCE_URL"), os.Getenv("CONJUR_ACCOUNT"))

		req, err := http.NewRequest(
			"GET",
			variableURL,
			nil,
		)
		assert.NoError(t, err)

		transport := &http.Transport{Proxy: func(req *http.Request) (proxyURL *url.URL, err error) {
			proxyURL, err = http.ProxyFromEnvironment(req)
			if proxyURL == nil && err == nil {
				// Local environment
				proxyURL, err = url.Parse("http://localhost:1080")
			}

			return
		}}
		client := &http.Client{Transport: transport}
		resp, err := client.Do(req)
		assert.NoError(t, err)

		value, err := response.DataResponse(resp)
		assert.NoError(t, err)

		assert.Equal(t, "secret", string(value))
	})
}
