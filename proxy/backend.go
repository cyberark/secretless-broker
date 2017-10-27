package proxy

import (
	"fmt"
	"os"
	"strings"

	"github.com/kgilpin/secretless-pg/config"
	"github.com/kgilpin/secretless-pg/conjur"
)

var HostUsername = os.Getenv("CONJUR_AUTHN_LOGIN")
var HostAPIKey = os.Getenv("CONJUR_AUTHN_API_KEY")

type BackendConnection interface {
	Configure() (*config.BackendConfig, error)
}

/**
 * A Backend connection which is configured through static YAML metadata.
 */
type StaticBackendConnection struct {
	BackendConfig config.BackendConfig
}

func (self StaticBackendConnection) Configure() (*config.BackendConfig, error) {
	return &self.BackendConfig, nil
}

/**
 * A Backend connection which is configured via Conjur resources.
 */
type ConjurBackendConnection struct {
	Resource string
}

func (self ConjurBackendConnection) Configure() (*config.BackendConfig, error) {
	var err error
	var token *string
	var url string

	if HostUsername == "" {
		return nil, fmt.Errorf("CONJUR_AUTHN_LOGIN is not specified")
	}
	if HostAPIKey == "" {
		return nil, fmt.Errorf("CONJUR_AUTHN_API_KEY is not specified")
	}

	if token, err = conjur.Authenticate(HostUsername, HostAPIKey); err != nil {
		return nil, err
	}

	configuration := config.BackendConfig{}
	resourceTokens := strings.SplitN(self.Resource, ":", 3)
	baseName := strings.Join([]string{resourceTokens[0], "variable", resourceTokens[2]}, "/")
	if configuration.Username, err = conjur.Secret(fmt.Sprintf("%s/username", baseName), *token); err != nil {
		return nil, err
	}
	if configuration.Password, err = conjur.Secret(fmt.Sprintf("%s/password", baseName), *token); err != nil {
		return nil, err
	}
	if url, err = conjur.Secret(fmt.Sprintf("%s/url", baseName), *token); err != nil {
		return nil, err
	}

	// Form of url is : 'dbcluster.myorg.com:5432/reports'
	tokens := strings.SplitN(url, "/", 2)
	configuration.Address = tokens[0]
	if len(tokens) == 2 {
		configuration.Database = tokens[1]
	}
	configuration.Options = make(map[string]string)

	return &configuration, nil
}
