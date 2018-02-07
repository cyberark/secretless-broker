package provider

import (
	"fmt"

	vault "github.com/hashicorp/vault/api"
)

// VaultProvider provides data values from the Conjur vault.
type VaultProvider struct {
	name   string
	client *vault.Client
}

// NewVaultProvider constructs a VaultProvider. The API client is configured from
// environment variables.
func NewVaultProvider(name string) (provider Provider, err error) {
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

// Value obtains a value by id. Any secret which is stored in the vault is recognized.
// The datatype returned by Vault is interface{}, which makes it somewhat complex to obtain
// a specific data item from a Vault secret. Currently, the secret data in Vault is required
// to be a Map which contains a "password" entry.
//
// This restriction/convention will be revisited and revised soon.
//
// See https://github.com/conjurinc/secretless/issues/6
func (p VaultProvider) Value(id string) (value []byte, err error) {
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
	valueObj, ok = secret.Data["password"]
	if !ok {
		err = fmt.Errorf("HashiCorp Vault provider expects the secret '%s' to contain a 'password' field", id)
		return
	}

	switch v := valueObj.(type) {
	case string:
		value = []byte(v)
	case []byte:
		value = v
	default:
		err = fmt.Errorf("HashiCorp Vault provider expects the 'password' to be a string or byte[], got %T", v)
	}
	return
}
