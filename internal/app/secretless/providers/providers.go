package providers

import (
	awsProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/awssecrets"
	conjurProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/conjur"
	envProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/env"
	fileProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/file"
	keychainProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/keychain"
	kubernetesProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/kubernetessecrets"
	literalProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/literal"
	vaultProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/vault"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// ProviderFactories contains the list of built-in provider factories
var ProviderFactories = map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error){
	"aws":        awsProvider.ProviderFactory,
	"conjur":     conjurProvider.ProviderFactory,
	"env":        envProvider.ProviderFactory,
	"file":       fileProvider.ProviderFactory,
	"keychain":   keychainProvider.ProviderFactory,
	"kubernetes": kubernetesProvider.ProviderFactory,
	"literal":    literalProvider.ProviderFactory,
	"vault":      vaultProvider.ProviderFactory,
}
