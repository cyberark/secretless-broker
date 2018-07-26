package conjur

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"

	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
)

// Provider provides data values from the Conjur vault.
type Provider struct {
	Name   string
	Conjur *conjurapi.Client
	Config conjurapi.Config

	Username string
	APIKey   string
}

func hasField(field string, params *map[string]string) (ok bool) {
	_, ok = (*params)[field]
	return
}

// ProviderFactory constructs a Conjur Provider. The API client is configured from
// environment variables.
func ProviderFactory(options plugin_v1.ProviderOptions) plugin_v1.Provider {
	config := conjurapi.LoadConfig()

	var conjur *conjurapi.Client
	var username, apiKey, tokenFile string
	var err error

	username = os.Getenv("CONJUR_AUTHN_LOGIN")
	apiKey = os.Getenv("CONJUR_AUTHN_API_KEY")
	tokenFile = os.Getenv("CONJUR_AUTHN_TOKEN_FILE")

	if username != "" && apiKey != "" {
		if conjur, err = conjurapi.NewClientFromKey(config, authn.LoginPair{username, apiKey}); err != nil {
			log.Fatalf("ERROR: Could not create new Conjur provider: %s", err)
		}
	} else if tokenFile != "" {
		if conjur, err = conjurapi.NewClientFromTokenFile(config, tokenFile); err != nil {
			log.Fatalf("ERROR: Could not create new Conjur provider: %s", err)
		}
	} else {
		log.Fatalln("ERROR: Unable to construct a Conjur provider client from the available credentials")
	}

	return &Provider{
		Name:     options.Name,
		Conjur:   conjur,
		Config:   config,
		Username: username,
		APIKey:   apiKey,
	}
}

// GetName returns the name of the provider
func (p Provider) GetName() string {
	return p.Name
}

// GetValue obtains a value by ID. The recognized IDs are:
//	* "accessToken"
// 	* Any Conjur variable ID
func (p Provider) GetValue(id string) ([]byte, error) {
	if id == "accessToken" {
		if p.Username != "" && p.APIKey != "" {
			// TODO: Use a cached access token from the client, once it's exposed
			return p.Conjur.Authenticate(authn.LoginPair{
				p.Username,
				p.APIKey,
			})
		}
		return nil, fmt.Errorf("Sorry, can't currently provide an accessToken unless username and apiKey credentials are provided")
	}

	tokens := strings.SplitN(id, ":", 3)
	switch len(tokens) {
	case 1:
		tokens = []string{p.Config.Account, "variable", tokens[0]}
	case 2:
		tokens = []string{p.Config.Account, tokens[0], tokens[1]}
	}

	return p.Conjur.RetrieveSecret(strings.Join(tokens, ":"))
}
