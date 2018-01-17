package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi/response"

	. "github.com/smartystreets/goconvey/convey"
)

// TestConjur_Handler verifies that Conjur API requests which are proxied through the Secretless
// handler do not require authentication credentials.
func TestConjur_Handler(t *testing.T) {
	conjurConfig := LoadTestConjurConfig()

	Convey("Can fetch a variable value", t, func() {
		variableURL := fmt.Sprintf("%s/secrets/%s/variable/db/password", conjurConfig.URL, conjurConfig.Account)

		req, err := http.NewRequest(
			"GET",
			variableURL,
			nil,
		)
		So(err, ShouldBeNil)

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
		So(err, ShouldBeNil)

		value, err := response.DataResponse(resp)
		So(err, ShouldBeNil)

		So(string(value), ShouldEqual, "secret")
	})
}
