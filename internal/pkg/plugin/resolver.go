package plugin

import (
	"fmt"
	"log"
	"strings"
	"sync"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
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

// Provider finds or creates a named provider.
func (resolver *Resolver) Provider(name string) (provider plugin_v1.Provider, err error) {
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
	provider, err = providerFactory(providerOptions)
	if err != nil {
		return nil, err
	}

	resolver.Providers[name] = provider

	return
}

// Resolve accepts an list of Providers and a list of Credentials and
// attempts to obtain the value of each Credential from the appropriate Provider.
func (resolver *Resolver) Resolve(credentials []*config_v2.Credential) (map[string][]byte, error) {
	if len(credentials) == 0 {
		resolver.LogFatalf("ERROR! Credentials not defined in Resolve call!")
	}

	result := make(map[string][]byte)
	errorStrings := make([]string, 0, len(credentials))

	var err error
	for _, credential := range credentials {
		var provider plugin_v1.Provider
		var value []byte

		if provider, err = resolver.Provider(credential.From); err != nil {
			resolver.LogFatalf("ERROR: Provider '%s' could not be used! %v", credential.From, err)
		}

		// This provider cannot resolve the named credential
		if value, err = provider.GetValue(credential.Get); err != nil {
			errInfo := fmt.Sprintf("ERROR: Resolving credential '%s' from provider '%s' failed: %v",
				credential.Get,
				credential.From,
				err)
			log.Println(errInfo)

			errorStrings = append(errorStrings, errInfo)
			continue
		}

		result[credential.Name] = value

		if resolver.EventNotifier != nil {
			resolver.EventNotifier.ResolveCredential(provider, credential.Name, value)
		}
	}

	err = nil
	if len(errorStrings) > 0 {
		err = fmt.Errorf(strings.Join(errorStrings, "\n"))
	}

	return result, err
}
