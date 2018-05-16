package provider

import (
	"io/ioutil"

	"github.com/conjurinc/secretless/pkg/secretless"
)

// FileProvider reads the contents of the specified file.
type FileProvider struct {
	name string
}

// NewFileProvider constructs a FileProvider.
// No configuration or credentials are required.
func NewFileProvider(name string) (provider secretless.Provider, err error) {
	provider = &FileProvider{name: name}

	return
}

// Name returns the name of the provider
func (p FileProvider) Name() string {
	return p.name
}

// Value reads the contents of the identified file.
func (p FileProvider) Value(id string) ([]byte, error) {
	return ioutil.ReadFile(id)
}
