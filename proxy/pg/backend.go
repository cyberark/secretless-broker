package pg

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"

	"github.com/kgilpin/secretless-pg/conjur"
)

var HostUsername = os.Getenv("CONJUR_AUTHN_LOGIN")
var HostAPIKey = os.Getenv("CONJUR_AUTHN_API_KEY")

type Variable interface {
	Value() (string, error)
}

/**
 * A variable which is provided as a Value literal.
 */
type ValueVariable struct {
	Literal string
}

func (self ValueVariable) Value() (string, error) {
	return self.Literal, nil
}

/**
 * A variable which is provided as a Conjur resource id.
 */
type ConjurVariable struct {
	Resource string
}

func (self ConjurVariable) Value() (string, error) {
	var err error
	var token *string

	if HostUsername == "" {
		return "", fmt.Errorf("CONJUR_AUTHN_LOGIN is not specified")
	}
	if HostAPIKey == "" {
		return "", fmt.Errorf("CONJUR_AUTHN_API_KEY is not specified")
	}

	if token, err = conjur.Authenticate(HostUsername, HostAPIKey); err != nil {
		return "", err
	}

	return conjur.Secret(self.Resource, *token)
}

/**
 * A variable which is provided as a file name.
 */
type FileVariable struct {
	File string
}

func (self FileVariable) Value() (string, error) {
	if bytes, err := ioutil.ReadFile(self.File); err != nil {
		return "", err
	} else {
		return string(bytes), nil
	}
}

func (self *PGHandler) ConfigureBackend() error {
	result := PGBackendConfig{Options: make(map[string]string)}

	for _, v := range self.Config.Backend {
		var variable Variable

		if v.Value != "" {
			variable = ValueVariable{v.Value}
		} else if v.ValueFrom.Conjur != "" {
			variable = ConjurVariable{v.ValueFrom.Conjur}
		} else if v.ValueFrom.File != "" {
			variable = FileVariable{v.ValueFrom.File}
		}
		if variable != nil {
			if value, err := variable.Value(); err != nil {
				return err
			} else {
				switch v.Name {
				case "address":
					// Form of url is : 'dbcluster.myorg.com:5432/reports'
					tokens := strings.SplitN(value, "/", 2)
					result.Address = tokens[0]
					if len(tokens) == 2 {
						result.Database = tokens[1]
					}
				case "username":
					result.Username =value
				case "password":
					result.Password = value
				default:
					result.Options[v.Name] = value
				}
			}
		}
	}

	self.BackendConfig = &result

	return nil
}
