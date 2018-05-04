package provider

import (
	"fmt"
	"os"
	"strings"

	"github.com/conjurinc/secretless/pkg/secretless"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
)

// ConjurProvider provides data values from the Conjur vault.
type ConjurProvider struct {
	name   string
	conjur *conjurapi.Client
	config conjurapi.Config

	username string
	apiKey   string
}

func hasField(field string, params *map[string]string) (ok bool) {
	_, ok = (*params)[field]
	return
}

// NewConjurProvider constructs a ConjurProvider. The API client is configured from
// environment variables.
func NewConjurProvider(name string) (provider secretless.Provider, err error) {
	config := conjurapi.LoadConfig()

	var conjur *conjurapi.Client
	var username, apiKey, tokenFile string

	username = os.Getenv("CONJUR_AUTHN_LOGIN")
	apiKey = os.Getenv("CONJUR_AUTHN_API_KEY")
	tokenFile = os.Getenv("CONJUR_AUTHN_TOKEN_FILE")

	if username != "" && apiKey != "" {
		conjur, err = conjurapi.NewClientFromKey(config, authn.LoginPair{username, apiKey})
	} else if tokenFile != "" {
		conjur, err = conjurapi.NewClientFromTokenFile(config, tokenFile)
	} else {
		err = fmt.Errorf("Unable to construct a Conjur API client from the available credentials")
	}

	if err != nil {
		return
	}

	provider = &ConjurProvider{name: name, conjur: conjur, config: config, username: username, apiKey: apiKey}

	return
}

// Name returns the name of the provider
func (p ConjurProvider) Name() string {
	return p.name
}

// Value obtains a value by ID. The recognized IDs are:
//	* "accessToken"
// 	* Any Conjur variable ID
func (p ConjurProvider) Value(id string) ([]byte, error) {
	if id == "accessToken" {
		if p.username != "" && p.apiKey != "" {
			// TODO: Use a cached access token from the client, once it's exposed
			return p.conjur.Authenticate(authn.LoginPair{p.username, p.apiKey})
		}
		return nil, fmt.Errorf("Sorry, can't currently provide an accessToken unless username and apiKey credentials are provided")
	}

	tokens := strings.SplitN(id, ":", 3)
	switch len(tokens) {
	case 1:
		tokens = []string{p.config.Account, "variable", tokens[0]}
	case 2:
		tokens = []string{p.config.Account, tokens[0], tokens[1]}
	}

	return p.conjur.RetrieveSecret(strings.Join(tokens, ":"))
}
