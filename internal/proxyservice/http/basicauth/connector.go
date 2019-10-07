package basicauth

import (
	"encoding/base64"
	"fmt"
	gohttp "net/http"
	"strconv"
	"strings"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
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
	var ok bool

	// Validate credentials

	// Ensure the needed credentials exist
	username, ok := credentialsByID["username"]
	if !ok {
		return fmt.Errorf("credential 'username' is not available")
	}
	password, ok := credentialsByID["password"]
	if !ok {
		return fmt.Errorf("credential 'password' is not available")
	}

	// Ensure username is valid
	if strings.Contains(string(username), ":") {
		return fmt.Errorf("'username' cannot contain a colon")
	}

	// Fulfill SSL requests

	forceSSLStr, ok := credentialsByID["forceSSL"]
	forceSSL, err := strconv.ParseBool(string(forceSSLStr))
	if ok && err == nil && forceSSL {
		r.URL.Scheme = "https"
	}

	// Add auth header to request

	rawAuthString := username
	rawAuthString = append(rawAuthString, []byte(":")...)
	rawAuthString = append(rawAuthString, password...)

	r.Header.Set(
		"Authorization",
		fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(rawAuthString)),
	)

	return nil
}
