package variable

import (
	providerPkg "github.com/conjurinc/secretless/internal/pkg/provider"
	"github.com/conjurinc/secretless/pkg/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
)

// Resolve accepts an list of Providers and a list of Variables and
// attempts to obtain the value of each Variable from the appropriate Provider.
func Resolve(variables []config.Variable) (result map[string][]byte, err error) {
	result = make(map[string][]byte)

	for _, v := range variables {
		var provider secretless.Provider
		var value []byte

		if provider, err = providerPkg.GetProvider(v.Provider); err != nil {
			return
		}

		if value, err = provider.Value(v.ID); err != nil {
			return
		}
		result[v.Name] = value
	}

	return
}
