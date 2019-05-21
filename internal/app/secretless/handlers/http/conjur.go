package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
)

// ConjurHandler applies Conjur authentication to the HTTP Authorization header.
type ConjurHandler struct {
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
