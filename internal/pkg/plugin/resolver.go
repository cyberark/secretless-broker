package plugin

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// Resolver is used toinstantiate providers and resolve credentials
type Resolver struct {
	EventNotifier     plugin_v1.EventNotifier
	ProviderFactories map[string]func(plugin_v1.ProviderOptions) plugin_v1.Provider
	Providers         map[string]plugin_v1.Provider
	LogFatalf         func(string, ...interface{})
}

func NewResolver(providerFactories map[string]func(plugin_v1.ProviderOptions) plugin_v1.Provider,
	eventNotifier plugin_v1.EventNotifier,
	LogFatalFunc func(string, ...interface{})) plugin_v1.Resolver {

	if LogFatalFunc == nil {
		LogFatalFunc = log.Fatalf
	}

	return &Resolver{
		EventNotifier:     eventNotifier,
		LogFatalf:         LogFatalFunc,
		ProviderFactories: providerFactories,
		Providers:         make(map[string]plugin_v1.Provider),
	}
}

// GetProvider finds or creates a named provider.
func (resolver *Resolver) GetProvider(name string) (provider plugin_v1.Provider, err error) {
	mutex := &sync.Mutex{}

	mutex.Lock()
	defer mutex.Unlock()

	if provider = resolver.Providers[name]; provider != nil {
		return
	}

	// If we don't know what this provider is, it's a critical error
	if _, ok := resolver.ProviderFactories[name]; !ok {
		resolver.LogFatalf("ERROR: Provider '%s' cannot be found", name)
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
func (resolver *Resolver) Resolve(variables []config.Variable) (map[string][]byte, error) {
	if variables == nil {
		resolver.LogFatalf("ERROR! Variables not defined in Resolve call!")
	}

	result := make(map[string][]byte)
	errorStrings := make([]string, 0, len(variables))

	var err error
	for _, variable := range variables {
		var provider plugin_v1.Provider
		var value []byte

		if provider, err = resolver.GetProvider(variable.Provider); err != nil {
			resolver.LogFatalf("ERROR: Provider '%s' could not be used! %v", variable.Provider, err)
		}

		// This provider cannot resolve the named variable
		if value, err = provider.GetValue(variable.ID); err != nil {
			errInfo := fmt.Sprintf("ERROR: Resolving variable '%s' from provider '%s' failed: %v",
				variable.ID,
				variable.Provider,
				err)
			log.Println(errInfo)

			errorStrings = append(errorStrings, errInfo)
			continue
		}

		result[variable.Name] = value

		if resolver.EventNotifier != nil {
			resolver.EventNotifier.ResolveVariable(provider, variable.Name, value)
		}
	}

	err = nil
	if len(errorStrings) > 0 {
		err = fmt.Errorf(strings.Join(errorStrings, "\n"))
	}

	return result, err
}
