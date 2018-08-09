package http

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"golang.org/x/crypto/ssh/agent"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// ConjurHandler applies Conjur authentication to the HTTP Authorization header.
type ConjurHandler struct {
	HandlerConfig config.Handler
	Resolver      plugin_v1.Resolver
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

// GetConfig implements secretless.Handler
func (h *ConjurHandler) GetConfig() config.Handler {
	return h.HandlerConfig
}

// GetClientConnection implements secretless.Handler
func (h *ConjurHandler) GetClientConnection() net.Conn {
	return nil
}

// GetBackendConnection implements secretless.Handler
func (h *ConjurHandler) GetBackendConnection() net.Conn {
	return nil
}

// LoadKeys is unused here
// TODO: Remove this when interface is cleaned up
func (h *ConjurHandler) LoadKeys(keyring agent.Agent) error {
	return errors.New("http/conjur handler does not use LoadKeys")
}

// ConjurHandlerFactory instantiates a handler given HandlerOptions
func ConjurHandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	return &ConjurHandler{
		HandlerConfig: options.HandlerConfig,
		Resolver:      options.Resolver,
	}
}
