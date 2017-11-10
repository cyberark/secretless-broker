package variable

import (
  "github.com/kgilpin/secretless/pkg/secretless/config"
)

func Resolve(variables []config.Variable) (*map[string]string, error) {
  result := make(map[string]string)

  for _, v := range variables {
    var variable Variable

    if v.Value != "" {
      variable = ValueVariable{v.Value}
    } else if v.ValueFrom.Conjur != "" {
      variable = ConjurVariable{v.ValueFrom.Conjur}
    } else if v.ValueFrom.Environment != "" {
      variable = EnvironmentVariable{v.ValueFrom.Environment}
    } else if v.ValueFrom.File != "" {
      variable = FileVariable{v.ValueFrom.File}
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
