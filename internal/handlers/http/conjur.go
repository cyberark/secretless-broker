package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
)

// ConjurHandler applies Conjur authentication to the HTTP Authorization header.
type ConjurHandler struct {
	plugin_v1.BaseHandler
}

// Authenticate applies the "accessToken" credential to the Authorization header, following the
// Conjur format:
//   Token token="<base64(accessToken)>"
func (h ConjurHandler) Authenticate(values map[string][]byte, r *http.Request) error {
	var ok bool

	accessToken, ok := values["accessToken"]
	if !ok {
		return fmt.Errorf("Conjur credential 'accessToken' is not available")
	}

	forceSSLStr, ok := values["forceSSL"]
	forceSSL, err := strconv.ParseBool(string(forceSSLStr))
	if ok && err == nil && forceSSL {
		r.URL.Scheme = "https"
	}

	r.Header.Set("Authorization", fmt.Sprintf("Token token=\"%s\"", base64.StdEncoding.EncodeToString(accessToken)))

	return nil
}

// ConjurHandlerFactory instantiates a handler given HandlerOptions
func ConjurHandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &ConjurHandler{
		BaseHandler: plugin_v1.NewBaseHandler(options),
	}

	// Force instantiate the Conjur provider so we can use an access token.
	// This will fail unless a means of authentication to Conjur is available.
	if handler.Resolver != nil {
		handler.Resolver.Provider("conjur");
	}

	return handler
}
