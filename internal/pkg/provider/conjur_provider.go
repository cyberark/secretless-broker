package provider

import (
  "fmt"
  "log"
  "strings"

  "github.com/cyberark/conjur-api-go/conjurapi"
  "github.com/cyberark/conjur-api-go/conjurapi/authn"
)

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

func NewConjurProvider(name string, configuration, credentials map[string]string) (provider Provider, err error) {
  config := conjurapi.Config{}
  for k, v := range configuration {
    switch k {
      case "url":
        config.ApplianceURL = v
      case "account":
        config.Account = v
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

  /*
  } else if hasField("token", credentials) {
    token := credentials["token"]
    conjur = conjurapi.NewClientFromToken(config, token)
  }
  */

  provider = &ConjurProvider{name: name, conjur: conjur, username: username, apiKey: apiKey}

  return
}

func (self ConjurProvider) Name() string {
  return self.name
}

func (self ConjurProvider) Value(id string) ([]byte, error) {
  if id == "accessToken" {
    if self.username != "" && self.apiKey != "" {
      // TODO: Use a cached access token from the client, once it's exposed
      return self.conjur.Authenticate(authn.LoginPair{self.username, self.apiKey})
    } else {
      return nil, fmt.Errorf("Sorry, can't currently provide an accessToken unless username and apiKey credentials are provided")
    }
  }

  tokens := strings.SplitN(id, ":", 3)
  switch len(tokens) {
  case 1:
    return nil, fmt.Errorf("%s does not know how to provide a value for '%s'", self.Name, id)
  case 2:
    if tokens[0] != "variable" {
      return nil, fmt.Errorf("%s does not know how to provide a value for '%s'", self.Name, id)
    }
    // TODO: change this to pass the full id, once the API client knows how to handle one.
    // tokens = []string{ conjur.Config.Account, tokens[0], tokens[1] }
    tokens = []string{ tokens[1] }
  }

  return self.conjur.RetrieveSecret(strings.Join(tokens, ":"))
}
