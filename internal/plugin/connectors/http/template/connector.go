package template

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"

	gohttp "net/http"
)

// Connector injects an HTTP request with AWS authorization headers.
type Connector struct {
	logger log.Logger
}

// Connect implements the http.Connector func signature.
func (c *Connector) Connect(
	r *gohttp.Request,
	credentialsByID connector.CredentialValuesByID,
) error {
	// TODO: add logic according to
	//  https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md#http-connector

	return nil
}
