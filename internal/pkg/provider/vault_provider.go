package provider

import (
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
)

// VaultProvider provides data values from the Conjur vault.
type VaultProvider struct {
	name   string
	client *vault.Client
}

// NewVaultProvider constructs a VaultProvider. configuration may include the following
// keys:
//   * address
//   * ca_cert
//   * client_cert
//   * tls_server_name
//
// credentials must contain the following:
//   * client_key
//   * token
func NewVaultProvider(name string, configuration, credentials map[string]string) (provider Provider, err error) {
	config := vault.DefaultConfig()

	var tls bool
	var token string
	tlsConfig := vault.TLSConfig{}
	for k, v := range configuration {
		switch k {
		case "address":
			config.Address = v
		case "ca_cert":
			tls = true
			tlsConfig.CACert = v
		case "client_cert":
			tls = true
			tlsConfig.ClientCert = v
		case "tls_server_name":
			tls = true
			tlsConfig.TLSServerName = v
		default:
			log.Printf("Unrecognized configuration setting '%s' for Hashicorp Vault provider %s", k, name)
		}
	}

	for k, v := range credentials {
		switch k {
		case "client_key":
			tls = true
			tlsConfig.ClientKey = v
		case "token":
			token = v
		default:
			log.Printf("Unrecognized credential '%s' for Hashicorp Vault provider %s", k, name)
		}
	}

	if tls {
		config.ConfigureTLS(&tlsConfig)
	}

	if token == "" {
		err = fmt.Errorf("Hashicorp Vault provider requires 'token' credential")
		return
	}

	var client *vault.Client

	if client, err = vault.NewClient(config); err != nil {
		return
	}
	client.SetToken(token)

	return VaultProvider{name: name, client: client}, nil
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
