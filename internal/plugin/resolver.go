package plugin

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
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
func NewResolver(
	providerFactories map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error),
	eventNotifier plugin_v1.EventNotifier,
	LogFatalFunc func(string, ...interface{}),
) plugin_v1.Resolver {

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

func (resolver *Resolver) resolveForProvider(
	providerID string,
	credentials []*config_v2.Credential,
) (plugin_v1.Provider, map[string]plugin_v1.ProviderResponse, string) {
	provider, err := resolver.Provider(providerID)
	if err != nil {
		resolver.LogFatalf("ERROR: Provider '%s' could not be used! %v", providerID, err)
	}

	// Create secretIds slice and credentialBySecretId map
	secretIds := make([]string, len(credentials))
	for idx, cred := range credentials {
		secretIds[idx] = cred.Get
	}

	// Resolves all credentials for current provider
	var hasErrors bool
	var providerErrStrings []string
	providerResponses, err := provider.GetValues(secretIds...)
	if err != nil {
		hasErrors = true
		providerErrStrings = append(
			providerErrStrings,
			err.Error(),
		)
	}

	// Collect errors from provider responses
	for _, providerResponse := range providerResponses {
		if providerResponse.Error != nil {
			hasErrors = true
			providerErrStrings = append(
				providerErrStrings,
				providerResponse.Error.Error(),
			)
			continue
		}
	}
	if hasErrors {
		// Sort the error strings to provide deterministic output
		sort.Strings(providerErrStrings)

		errInfo := fmt.Sprintf(
			"ERROR: Resolving credentials from provider '%s' failed: %v",
			provider.GetName(),
			strings.Join(providerErrStrings, ", "),
		)
		return nil, nil, errInfo
	}

	return provider, providerResponses, ""
}

// Resolve accepts n list of Providers and a list of Credentials and
// attempts to obtain the value of each Credential from the appropriate Provider.
func (resolver *Resolver) Resolve(credentials []*config_v2.Credential) (map[string][]byte, error) {
	if len(credentials) == 0 {
		resolver.LogFatalf("ERROR! Credentials not defined in Resolve call!")
	}

	result := make(map[string][]byte)
	errorStrings := make([]string, 0, len(credentials))

	var err error

	// Group credentials by provider
	credentialsByProvider := groupCredentialsByProvider(credentials)

	// Resolve credentials by provider
	for providerID, credentialsForProvider := range credentialsByProvider {
		provider, providerResponses, errStr := resolver.resolveForProvider(
			providerID,
			credentialsForProvider,
		)
		if errStr != "" {
			log.Println(errStr)

			errorStrings = append(errorStrings, errStr)
			continue
		}

		// Set provider responses on result map before returning
		for _, credential := range credentialsForProvider {
			credentialName := credential.Name
			secretValue := providerResponses[credential.Get].Value

			result[credentialName] = secretValue

			if resolver.EventNotifier != nil {
				resolver.EventNotifier.ResolveCredential(
					provider,
					credentialName,
					secretValue,
				)
			}
		}
	}

	err = nil
	if len(errorStrings) > 0 {
		// Sort the error strings to provide deterministic output
		sort.Strings(errorStrings)

		err = fmt.Errorf(strings.Join(errorStrings, "\n"))
	}

	return result, err
}

func groupCredentialsByProvider(
	credentials []*config_v2.Credential,
) map[string][]*config_v2.Credential {
	credentialsByProvider := make(map[string][]*config_v2.Credential)
	for _, credential := range credentials {
		credentialsByProvider[credential.From] = append(
			credentialsByProvider[credential.From],
			credential,
		)
	}

	return credentialsByProvider
}
