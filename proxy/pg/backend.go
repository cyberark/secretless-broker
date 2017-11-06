package pg

import (
  "strings"

  "github.com/kgilpin/secretless/variable"
)

func (self *Handler) ConfigureBackend() error {
  result := BackendConfig{Options: make(map[string]string)}

  if valuesPtr, err := variable.Resolve(self.Config.Backend); err != nil {
    return err
  } else {
  	values := *valuesPtr
    if address := values["address"]; address != "" {
      // Form of url is : 'dbcluster.myorg.com:5432/reports'
      tokens := strings.SplitN(address, "/", 2)
      result.Address = tokens[0]
      if len(tokens) == 2 {
        result.Database = tokens[1]
      }
    }

    result.Username = values["username"]
    result.Password = values["password"]

    delete(values, "address")
    delete(values, "username")
    delete(values, "password")

    result.Options = values
  }

  self.BackendConfig = &result

  return nil
}
