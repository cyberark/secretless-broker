package template

import (
	gohttp "net/http"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Connector injects an HTTP request with AWS authorization headers.
type Connector struct {
	logger log.Logger
}

/*
	This function has access to the client http.Request and the credentials
	(as a map), and is expected to modify the request so that it will authenticate.
	This typically means adding required authorization headers.
*/
func (c *Connector) Connect(
	r *gohttp.Request,
	credentialsByID connector.CredentialValuesByID,
) error {
	// TODO: add logic according to
	// https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md#http-connector
	// http/basicauth/connector.go is a good example.

	var err error
	return err
}
