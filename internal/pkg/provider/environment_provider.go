package provider

import (
	"fmt"
	"os"
)

// EnvironmentProvider provides data values from the process environment.
type EnvironmentProvider struct {
	name string
}

// NewEnvironmentProvider constructs a EnvironmentProvider.
// No configuration or credentials are required.
func NewEnvironmentProvider(name string) (provider Provider, err error) {
	provider = &EnvironmentProvider{name: name}

	return
}

// Name returns the name of the provider
func (p EnvironmentProvider) Name() string {
	return p.name
}

// Value obtains a value by ID. The recognized IDs are:
//	* "accessToken"
// 	* Any Environment variable ID
func (p EnvironmentProvider) Value(id string) (result []byte, err error) {
	var found bool
	envVar, found := os.LookupEnv(id)
	if found {
		result = []byte(envVar)
	} else {
		err = fmt.Errorf("%s cannot find environment variable '%s'", p.Name(), id)
	}
	return
}
