package handler

import (
  "encoding/base64"
  "fmt"
  "net/http"

  "github.com/kgilpin/secretless/pkg/secretless/config"
)

type ConjurHandler struct {
  Config config.Handler
}

func (self ConjurHandler) Configuration() *config.Handler {
  return &self.Config
}

func (self ConjurHandler) Authenticate(values map[string]string, r *http.Request) error {
	accessToken := values["accessToken"]
	if accessToken == "" {
		return fmt.Errorf("Conjur credential 'accessToken' is not available")
	}

  forceSSL, ok := values["forceSSL"]
  if ok && forceSSL == "true" {
    r.URL.Scheme = "https"
  }
  r.Header.Set("Authorization", fmt.Sprintf("Token token=\"%s\"", base64.StdEncoding.EncodeToString([]byte(accessToken))))

  return nil
}
