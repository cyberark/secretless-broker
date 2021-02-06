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

	// Extract metadata of a signed AWS request: Date, Region and service name
	reqMeta, err := NewRequestMetadata(req)
	if err != nil {
		return err
	}

	// No metadata means the original request was not signed. Don't sign this
	// request either.
	if reqMeta == nil {
		return nil
	}

	// Set AWS endpoint
	// NOTE: this must be done before signing the request, otherwise the modified request
	// will fail the integrity check.
	err = setAmzEndpoint(req, reqMeta)
	if err != nil {
		return err
	}

	// Use metadata and credentials to sign request
	c.logger.Debugf(
		"Signing for service=%s Region=%s",
		reqMeta.ServiceName,
		reqMeta.Region,
	)
	return signRequest(req, reqMeta, credentialsByID)
}
