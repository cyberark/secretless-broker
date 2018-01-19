package provider

import (
	"fmt"
	"log"
	"strings"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
)

// ConjurProvider provides data values from the Conjur vault.
type ConjurProvider struct {
	name   string
	conjur *conjurapi.Client

	username string
	apiKey   string
}

func hasField(field string, params *map[string]string) (ok bool) {
	_, ok = (*params)[field]
	return
}

// NewConjurProvider constructs a ConjurProvider. configuration may include the following
// keys:
//  * url (required)
//  * account (required)
//  * version (optional; 4 or 5)
//  * certFile (optional)
//
// credentials may contain the following:
//  * username (optional; requires apiKey)
//  * apiKey (optional; requires username)
//  * tokenFile (optional; username and apiKey are not required if this is present)
func NewConjurProvider(name string, configuration, credentials map[string]string) (provider Provider, err error) {
	config := conjurapi.Config{}
	for k, v := range configuration {
		switch k {
		case "url":
			config.ApplianceURL = v
		case "account":
			config.Account = v
		case "certFile":
			config.SSLCertPath = v
		case "version":
			if v == "4" {
				config.V4 = true
			}
		/* todo: the others, e.g. SSL */
		default:
			log.Printf("Unrecognized configuration setting '%s' for Conjur provider %s", k, name)
		}
	}

	var conjur *conjurapi.Client
	var username, apiKey string
	if hasField("username", &credentials) && hasField("apiKey", &credentials) {
		username = credentials["username"]
		apiKey = credentials["apiKey"]
		conjur, err = conjurapi.NewClientFromKey(config, authn.LoginPair{username, apiKey})
	} else if hasField("tokenFile", &credentials) {
		tokenFile := credentials["tokenFile"]
		conjur, err = conjurapi.NewClientFromTokenFile(config, tokenFile)
	} else {
		err = fmt.Errorf("Unable to construct a Conjur API client from the provided credentials")
	}

	if err != nil {
		return
	}

	provider = &ConjurProvider{name: name, conjur: conjur, username: username, apiKey: apiKey}

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
		return nil, fmt.Errorf("%s does not know how to provide a value for '%s'", p.Name(), id)
	case 2:
		if tokens[0] != "variable" {
			return nil, fmt.Errorf("%s does not know how to provide a value for '%s'", p.Name(), id)
		}
		// TODO: change this to pass the full id, once the API client knows how to handle one.
		// tokens = []string{ conjur.Config.Account, tokens[0], tokens[1] }
		tokens = []string{tokens[1]}
	}

	return p.conjur.RetrieveSecret(strings.Join(tokens, ":"))
}
