package providers

import (
	conjurProvider "github.com/conjurinc/secretless/internal/app/secretless/providers/conjur"
	envProvider "github.com/conjurinc/secretless/internal/app/secretless/providers/env"
	fileProvider "github.com/conjurinc/secretless/internal/app/secretless/providers/file"
	keychainProvider "github.com/conjurinc/secretless/internal/app/secretless/providers/keychain"
	literalProvider "github.com/conjurinc/secretless/internal/app/secretless/providers/literal"
	vaultProvider "github.com/conjurinc/secretless/internal/app/secretless/providers/vault"

	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
)

// ProviderFactories contains the list of built-in provider factories
var ProviderFactories = map[string]func(plugin_v1.ProviderOptions) plugin_v1.Provider{
	"conjur":   conjurProvider.ProviderFactory,
	"env":      envProvider.ProviderFactory,
	"file":     fileProvider.ProviderFactory,
	"keychain": keychainProvider.ProviderFactory,
	"literal":  literalProvider.ProviderFactory,
	"vault":    vaultProvider.ProviderFactory,
}
