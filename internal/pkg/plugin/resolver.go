package plugin

import (
	"log"
	"sync"

	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
)

// Resolver is used toinstantiate providers and resolve credentials
type Resolver struct {
	EventNotifier     plugin_v1.EventNotifier
	ProviderFactories map[string]func(plugin_v1.ProviderOptions) plugin_v1.Provider
	Providers         map[string]plugin_v1.Provider
}

// GetProvider finds or creates a named provider.
func (resolver *Resolver) GetProvider(name string) (provider plugin_v1.Provider, err error) {
	mutex := &sync.Mutex{}

	mutex.Lock()
	defer mutex.Unlock()

	if resolver.Providers == nil {
		resolver.Providers = make(map[string]plugin_v1.Provider)
	}

	if provider = resolver.Providers[name]; provider != nil {
		return
	}

	// If we don't know what this provider is, it's a critical error
	if _, ok := resolver.ProviderFactories[name]; !ok {
		log.Fatalf("ERROR: Provider '%s' cannot be found", name)
	}

	providerOptions := plugin_v1.ProviderOptions{
		Name: name,
	}

	providerFactory := resolver.ProviderFactories[name]

	log.Printf("Instantiating provider '%s'", name)
	provider = providerFactory(providerOptions)

	resolver.Providers[name] = provider

	return
}

// Resolve accepts an list of Providers and a list of Variables and
// attempts to obtain the value of each Variable from the appropriate Provider.
func (resolver *Resolver) Resolve(variables []config.Variable) (result map[string][]byte, err error) {
	if variables == nil {
		log.Fatalln("ERROR! Variables not defined in Resolve call!")
	}

	result = make(map[string][]byte)

	for _, variable := range variables {
		var provider plugin_v1.Provider
		var value []byte

		if provider, err = resolver.GetProvider(variable.Provider); err != nil {
			log.Fatalf("ERROR: Provider '%s' could not be used!", variable.Provider)
		}

		// This provider cannot resolve the named variable
		if value, err = provider.GetValue(variable.ID); err != nil {
			return
		}

		result[variable.Name] = value

		if resolver.EventNotifier != nil {
			resolver.EventNotifier.ResolveVariable(provider, variable.Name, value)
		}
	}

	return
}
