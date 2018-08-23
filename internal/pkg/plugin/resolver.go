package plugin

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// Resolver is used to instantiate providers and resolve credentials
type Resolver struct {
	EventNotifier     plugin_v1.EventNotifier
	ProviderFactories map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error)
	Providers         map[string]plugin_v1.Provider
	LogFatalf         func(string, ...interface{})
}

// NewResolver instantiates providers based on the name and ProviderOptions
func NewResolver(providerFactories map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error),
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
		return nil, fmt.Errorf("ERROR: Provider '%s' cannot be found", name)
	}

	providerOptions := plugin_v1.ProviderOptions{
		Name: name,
	}

	providerFactory := resolver.ProviderFactories[name]

	log.Printf("Instantiating provider '%s'", name)
	provider, err = providerFactory(providerOptions)
	if err != nil {
		return nil, err
	}

	resolver.Providers[name] = provider

	return
}

// Resolve accepts an list of Providers and a list of Variables and
// attempts to obtain the value of each Variable from the appropriate Provider.
func (resolver *Resolver) Resolve(variables []config.Variable) (map[string][]byte, error) {
	if variables == nil {
		return nil, fmt.Errorf("resolver received empty slice of variable ids")
	}

	// Sort variables by provider for batch retrieval
	sortedVariables := map[plugin_v1.Provider][]config.Variable{}
	for _, variable := range variables {
		var err error
		var provider plugin_v1.Provider
		if provider, err = resolver.GetProvider(variable.Provider); err != nil {
			resolver.LogFatalf("ERROR: Provider '%s' could not be used! %v", variable.Provider, err)
		}

		if slice, ok := sortedVariables[provider]; ok {
			sortedVariables[provider] = append(slice, variable)
		} else {
			sortedVariables[provider] = []config.Variable{variable}
		}
	}

	resultMutex := &sync.Mutex{}
	errorMutex := &sync.Mutex{}
	threads := &sync.WaitGroup{}
	threads.Add(len(sortedVariables))
	result := map[string][]byte{}
	errorStrings := make([]string, 0, len(variables))
	for provider, providerVariables := range sortedVariables {
		_variables := providerVariables
		_provider := provider

		go func() {
			values, err := _provider.GetValues(_variables)
			if err != nil {
				errInfo := fmt.Sprintf("Failed to resolve variable: %v", err)
				errorMutex.Lock()
				errorStrings = append(errorStrings, errInfo)
				errorMutex.Unlock()
				log.Println(errInfo)
			}

			resultMutex.Lock()
			for k, v := range values {
				result[k] = v
			}
			resultMutex.Unlock()

			threads.Done()
		}()
	}

	threads.Wait();

	var err error
	if len(errorStrings) > 0 {
		err = fmt.Errorf(strings.Join(errorStrings, "\n"))
	}

	return result, err
}