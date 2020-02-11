package aws

import (
	gohttp "net/http"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Connector injects an HTTP request with AWS authorization headers.
type Connector struct {
	logger log.Logger
}

// Connect is the function that implements the http.Connector func
// signature. It has access to the client http.Request and the credentials (as a
// map), and is expected to decorate the request with Authorization headers.
//
// Connect uses the "accessKeyId", "secretAccessKey" and optional "accessToken"
// credentials to sign the Authorization header, following the AWS signature
// format.
func (c *Connector) Connect(
	req *gohttp.Request,
	credentialsByID connector.CredentialValuesByID,
) error {
	var err error

	// Extract metadata of a signed AWS request: date, region and service name
	reqMeta, err := newRequestMetadata(req)
	if err != nil {
		return err
	}

	// No metadata means the original request was not signed. Don't sign this
	// request either.
	if reqMeta == nil {
		return nil
	}

	// Use metadata and credentials to sign request
	c.logger.Debugf(
		"Signing for service=%s region=%s",
		reqMeta.serviceName,
		reqMeta.region,
	)
	err = signRequest(req, reqMeta, credentialsByID)
	if err != nil {
		return err
	}

	// Set AWS endpoint
	return setAmzEndpoint(req, reqMeta)
}
