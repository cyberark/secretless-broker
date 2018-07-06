package variable

import (
	"log"

	providerPkg "github.com/conjurinc/secretless/internal/pkg/provider"
	"github.com/conjurinc/secretless/pkg/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
)

// Resolve accepts an list of Providers and a list of Variables and
// attempts to obtain the value of each Variable from the appropriate Provider.
func Resolve(variables []config.Variable, eventNotifier plugin_v1.EventNotifier) (result map[string][]byte, err error) {
	if variables == nil {
		log.Fatalln("ERROR! Variables not defined in Resolve call!")
	}

	if eventNotifier == nil {
		log.Fatalln("ERROR! EventNotifier not defined in Resolve call!")
	}

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

		eventNotifier.ResolveVariable(provider, v.Name, value)
	}

	return
}
