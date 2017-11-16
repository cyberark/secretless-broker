package variable

import (
  "fmt"
  "log"

  "github.com/kgilpin/secretless/pkg/secretless/config"
  "github.com/kgilpin/secretless/internal/pkg/provider"
)

func Resolve(providers []provider.Provider, variables []config.Variable) (*map[string]string, error) {
  result := make(map[string]string)

  for _, v := range variables {
    var variable Variable

    if v.Value.Literal != "" {
      variable = ValueVariable{v.Value.Literal}
    } else if v.Value.Provider != "" {
      var provider provider.Provider
      for i := range providers {
        if providers[i].Name() == v.Value.Provider {
          provider = providers[i]
          break
        }
      }
      if provider == nil {
        return nil, fmt.Errorf("Provider '%s' not found", v.Value.Provider)
      }
      variable = ProviderVariable{provider, v.Value.Id}
    } else if v.Value.Environment != "" {
      variable = EnvironmentVariable{v.Value.Environment}
    } else if v.Value.File != "" {
      variable = FileVariable{v.Value.File}
    } else if v.Value.Keychain.Service != "" {
      variable = KeychainVariable{v.Value.Keychain.Service, v.Value.Keychain.Username}
    }
    if variable != nil {
      if value, err := variable.Value(); err != nil {
        return nil, err
      } else {
        result[v.Name] = value
      }
    }
  }

  return &result, nil
}
