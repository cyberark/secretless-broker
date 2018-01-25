package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/conjurinc/secretless/pkg/secretless/config"
)

// ConjurHandler applies Conjur authentication to the HTTP Authorization header.
type ConjurHandler struct {
	Config config.Handler
}

// Configuration provides the handler configuration.
func (h ConjurHandler) Configuration() *config.Handler {
	return &h.Config
}

// Authenticate applies the "accessToken" credential to the Authorization header, following the
// Conjur format:
//   Token token="<base64(accessToken)>"
func (h ConjurHandler) Authenticate(values map[string]string, r *http.Request) error {
	var ok bool

	accessToken, ok := values["accessToken"]
	if !ok {
		return fmt.Errorf("Conjur credential 'accessToken' is not available")
	}

	forceSSLStr, ok := values["forceSSL"]
	forceSSL, err := strconv.ParseBool(forceSSLStr)
	if ok && err == nil && forceSSL {
		r.URL.Scheme = "https"
	}

	r.Header.Set("Authorization", fmt.Sprintf("Token token=\"%s\"", base64.StdEncoding.EncodeToString([]byte(accessToken))))

	return nil
}
