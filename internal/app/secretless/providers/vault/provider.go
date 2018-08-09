package vault

import (
	"fmt"
	"log"
	"strings"

	vault "github.com/hashicorp/vault/api"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// Provider provides data values from the Conjur vault.
type Provider struct {
	Name   string
	Client *vault.Client
}

// ProviderFactory constructs a Provider. The API client is configured from
// environment variables.
func ProviderFactory(options plugin_v1.ProviderOptions) plugin_v1.Provider {
	config := vault.DefaultConfig()

	var client *vault.Client
	var err error
	if client, err = vault.NewClient(config); err != nil {
		log.Panicf("ERROR: Could not create Vault provider: %s", err)
	}

	provider := Provider{
		Name:   options.Name,
		Client: client,
	}

	return provider
}

// GetName returns the name of the provider
func (p Provider) GetName() string {
	return p.Name
}

// VaultDefaultField is the default value returned by the provider from the
// hash returned by the Vault.
const VaultDefaultField = "value"

// parseVaultID returns the secret id and field name.
func parseVaultID(id string) (string, string) {
	tokens := strings.SplitN(id, "#", 2)
	switch len(tokens) {
	case 1:
		return tokens[0], VaultDefaultField
	default:
		return tokens[0], tokens[1]
	}
}

// GetValue obtains a value by id. Any secret which is stored in the vault is recognized.
// The datatype returned by Vault is map[string]interface{}. Therefore this provider needs
// to know which field to return from the map. By default, it returns the 'value'.
// An alternative field can be obtained by appending '#fieldName' to the id argument.
func (p Provider) GetValue(id string) (value []byte, err error) {
	id, fieldName := parseVaultID(id)

	var secret *vault.Secret
	if secret, err = p.Client.Logical().Read(id); err != nil {
		return
	}
	// secret can be nil if it's not found
	if secret == nil {
		err = fmt.Errorf("HashiCorp Vault provider could not find a secret called '%s'", id)
		return
	}

	var ok bool
	var valueObj interface{}
	valueObj, ok = secret.Data[fieldName]
	if !ok {
		err = fmt.Errorf("HashiCorp Vault provider expects the secret '%s' to contain field '%s'", id, fieldName)
		return
	}

	switch v := valueObj.(type) {
	case string:
		value = []byte(v)
	case []byte:
		value = v
	default:
		err = fmt.Errorf("HashiCorp Vault provider expects the secret to be a string or byte[], got %T", v)
	}
	return
}
