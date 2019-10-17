package conjur

import (
	"encoding/base64"
	"fmt"
	gohttp "net/http"
	"strconv"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// Connector injects an HTTP request with Conjur's authorization header(s).
type Connector struct {
	logger log.Logger
}

// Connect is the function that implements the http.Connector func
// signature. It has access to the client http.Request and the credentials (as a
// map), and is expected to decorate the request with Authorization headers.
//
// Connect uses the "accessToken" and "forceSSL" credentials to sign the Authorization
// header, following the Conjur format:
//   Token token="<base64(accessToken)>"
func (c *Connector) Connect(
	req *gohttp.Request,
	credentialsByID connector.CredentialValuesByID,
) error {
	var ok bool

	accessToken, ok := credentialsByID["accessToken"]
	if !ok {
		return fmt.Errorf("conjur credential 'accessToken' is not available")
	}

	forceSSLStr, ok := credentialsByID["forceSSL"]
	forceSSL, err := strconv.ParseBool(string(forceSSLStr))
	if ok && err == nil && forceSSL {
		req.URL.Scheme = "https"
	}

	// Use credentials to sign request
	c.logger.Debugln("Signing Conjur HTTP request...")

	req.Header.Set(
		"Authorization",
		fmt.Sprintf("Token token=\"%s\"", base64.StdEncoding.EncodeToString(accessToken)),
	)

	return nil
}
