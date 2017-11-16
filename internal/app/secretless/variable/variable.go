package variable

import (
  "io/ioutil"
  "log"
  "os"

  "github.com/kgilpin/secretless/internal/pkg/keychain"
  "github.com/kgilpin/secretless/internal/pkg/provider"
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
 * A variable which is provided by a configured provider.
 */
type ProviderVariable struct {
  Provider provider.Provider
  Id       string
}

func (self ProviderVariable) Value() (string, error) {
  value, err := self.Provider.Value(self.Id)
  if err != nil {
    return "", err
  }
  return string(value), nil
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

/**
 * A variable which is provided by cross-platform "Keychain" access.
 */
type KeychainVariable struct {
  Service string
  Username string
}

func (self KeychainVariable) Value() (string, error) {
  log.Printf("Loading API key from username '%s' in OS keychain service '%s'", self.Username, self.Service)
  return keychain.GetGenericPassword(self.Service, self.Username)
}

