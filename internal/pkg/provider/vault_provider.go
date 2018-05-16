package provider

import (
	"fmt"
	"strings"

	"github.com/conjurinc/secretless/pkg/secretless"
	vault "github.com/hashicorp/vault/api"
)

// VaultProvider provides data values from the Conjur vault.
type VaultProvider struct {
	name   string
	client *vault.Client
}

// NewVaultProvider constructs a VaultProvider. The API client is configured from
// environment variables.
func NewVaultProvider(name string) (provider secretless.Provider, err error) {
	config := vault.DefaultConfig()

	var client *vault.Client
	if client, err = vault.NewClient(config); err != nil {
		return
	}

	provider = VaultProvider{name: name, client: client}
	return
}

// Name returns the name of the provider
func (p VaultProvider) Name() string {
	return p.name
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

// Value obtains a value by id. Any secret which is stored in the vault is recognized.
// The datatype returned by Vault is map[string]interface{}. Therefore this provider needs
// to know which field to return from the map. By default, it returns the 'value'.
// An alternative field can be obtained by appending '#fieldName' to the id argument.
func (p VaultProvider) Value(id string) (value []byte, err error) {
	id, fieldName := parseVaultID(id)

	var secret *vault.Secret
	if secret, err = p.client.Logical().Read(id); err != nil {
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
