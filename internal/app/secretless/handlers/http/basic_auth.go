package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
)

// BasicAuthHandler applies HTTP Basic authentication to the HTTP Authorization header.
type BasicAuthHandler struct {
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
