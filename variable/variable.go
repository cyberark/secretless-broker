package variable

import (
  "io/ioutil"
  "os"

  "github.com/kgilpin/secretless/conjur"
)

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
 * A variable which is provided as a Value literal.
 */
type EnvironmentVariable struct {
  Literal string
}

func (self EnvironmentVariable) Value() (string, error) {
  return os.Getenv(self.Literal), nil
}

/**
 * A variable which is provided as a Conjur resource id.
 */
type ConjurVariable struct {
  Resource string
}

func (self ConjurVariable) Value() (string, error) {
  return conjur.Secret(self.Resource, conjur.AccessToken{UseDefault: true})
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
