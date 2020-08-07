package generic

import (
	"fmt"
	gohttp "net/http"

	oauth1 "github.com/cyberark/secretless-broker/internal/plugin/connectors/http/generic/oauth/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Connector injects an HTTP requests with authorization headers.
type Connector struct {
	logger log.Logger
	config *config
}

func addOAuth1Header(c *Connector, credentialsByID connector.CredentialValuesByID, r *gohttp.Request) error {
	// Doesn't check for error. Instead we check to see if
	// any values exist in the oauth1 key in the config.
	oAuth1Params, err := renderTemplates(c.config.OAuth1Secrets, credentialsByID)
	if err != nil {
		return fmt.Errorf("failed to render oauth1 params: %s", err)
	}

	if len(oAuth1Params) > 0 {
		oauthHeader, err := oauth1.CreateOAuth1Header(oAuth1Params, r)
		if err != nil {
			return fmt.Errorf("failed to create oAuth1 'Authorization' header: %s", err)
		}

		if len(r.Header.Get("Authorization")) > 0 {
			return fmt.Errorf("authorization header already exists, cannot override header")
		}

		r.Header.Set("Authorization", oauthHeader)
	}
	return nil
}

func addQueryParams(c *Connector, credentialsByID connector.CredentialValuesByID, r *gohttp.Request) error {
	queryParams, err := renderTemplates(c.config.QueryParams, credentialsByID)
	if err != nil {
		return fmt.Errorf("failed to render query params: %s", err)
	}
	r.URL.RawQuery = appendQueryParams(*r.URL, queryParams)
	return nil
}

func addHeaders(c *Connector, credentialsByID connector.CredentialValuesByID, r *gohttp.Request) error {
	headers, err := renderTemplates(c.config.Headers, credentialsByID)
	if err != nil {
		return fmt.Errorf("failed to render headers: %s", err)
	}
	for headerName, headerVal := range headers {
		r.Header.Set(headerName, headerVal)
	}
	return nil
}

// Connect implements the http.Connector func signature.
func (c *Connector) Connect(
	r *gohttp.Request,
	credentialsByID connector.CredentialValuesByID,
) error {
	// Validate credential values match expected patterns
	if err := c.config.validate(credentialsByID); err != nil {
		return err
	}

	// Fulfill SSL requests
	if c.config.ForceSSL {
		r.URL.Scheme = "https"
	}

	// Add configured headers to request
	if err := addHeaders(c, credentialsByID, r); err != nil {
		return err
	}

	// Add configured params to request
	if err := addQueryParams(c, credentialsByID, r); err != nil {
		return err
	}

	// Add oAuth1 Authorization Header to the request
	if err := addOAuth1Header(c, credentialsByID, r); err != nil {
		return err
	}

	return nil
}
