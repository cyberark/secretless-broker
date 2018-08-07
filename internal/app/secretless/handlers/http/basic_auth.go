package http

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"golang.org/x/crypto/ssh/agent"

	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
)

// BasicAuthHandler applies HTTP Basic authentication to the HTTP Authorization header.
type BasicAuthHandler struct {
	HandlerConfig config.Handler
	Resolver      plugin_v1.Resolver
}

// Authenticate applies the "username" and "password" credential to the Authorization header, following the
// RFC: Basic "<base64(<username> + ":" + <password>)>"
func (h BasicAuthHandler) Authenticate(values map[string][]byte, r *http.Request) error {
	var ok bool

	username, ok := values["username"]
	if !ok {
		return fmt.Errorf("Credential 'username' is not available")
	}
	// TODO: Check to ensure that username has no colons in the string

	password, ok := values["password"]
	if !ok {
		return fmt.Errorf("Credential 'password' is not available")
	}

	forceSSLStr, ok := values["forceSSL"]
	forceSSL, err := strconv.ParseBool(string(forceSSLStr))
	if ok && err == nil && forceSSL {
		r.URL.Scheme = "https"
	}

	rawAuthString := username
	rawAuthString = append(rawAuthString, []byte(":")...)
	rawAuthString = append(rawAuthString, password...)

	r.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(rawAuthString)))

	return nil
}

// GetConfig implements secretless.Handler
func (h *BasicAuthHandler) GetConfig() config.Handler {
	return h.HandlerConfig
}

// GetClientConnection implements secretless.Handler
func (h *BasicAuthHandler) GetClientConnection() net.Conn {
	return nil
}

// GetBackendConnection implements secretless.Handler
func (h *BasicAuthHandler) GetBackendConnection() net.Conn {
	return nil
}

// LoadKeys is unused here
// TODO: Remove this when interface is cleaned up
func (h *BasicAuthHandler) LoadKeys(keyring agent.Agent) error {
	return errors.New("http/conjur handler does not use LoadKeys")
}

func (h *BasicAuthHandler) Shutdown() error {
	return nil
}

// BasicAuthHandlerFactory instantiates a handler given HandlerOptions
func BasicAuthHandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	return &BasicAuthHandler{
		HandlerConfig: options.HandlerConfig,
		Resolver:      options.Resolver,
	}
}
