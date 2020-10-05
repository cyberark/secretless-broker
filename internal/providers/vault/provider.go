package vault

import (
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

// Provider provides data values from the Conjur vault.
type Provider struct {
	Name   string
	Client *vault.Client
}

// ProviderFactory constructs a Provider. The API client is configured from
// environment variables. Underlying Vault API client by default uses:
// - VAULT_ADDR: endpoint of Vault, e.g. http://vault:8200/
// - VAULT_TOKEN: token to login to Vault
// See Vault API docs at https://godoc.org/github.com/hashicorp/vault/api
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	config := vault.DefaultConfig()

	var client *vault.Client
	var err error
	if client, err = vault.NewClient(config); err != nil {
		return nil, fmt.Errorf("ERROR: Could not create Vault provider: %s", err)
	}

	provider := &Provider{
		Name:   options.Name,
		Client: client,
	}

	return provider, nil
}

// GetName returns the name of the provider
func (p *Provider) GetName() string {
	return p.Name
}

// DefaultField is the default field name the provider expects to find the secret value.
const DefaultField = "value"

// parseVaultID returns the path to the secret (object) and (normalized) field/property path to the secret value.
func parseVaultID(id string) (string, string) {
	tokens := strings.SplitN(id, "#", 2)
	switch len(tokens) {
	case 1:
		return tokens[0], DefaultField
	default:
		return tokens[0], tokens[1]
	}
}

// valueOf returns the value when navigating the obj along the fields.
// Suppose obj = { "foo": { "bar": "qux" } } and fields = "foo.bar", then value returned is "qux".
func valueOf(obj map[string]interface{}, fields string) (interface{}, bool) {
	// Split fields to navigate by ".", e.g. if fields = [ "foo.bar" ] then it becomes a slice of [ "foo", "bar" ]
	nav := strings.Split(fields, ".")

	// Traverse, starting at given obj, moving deeper into the object structure
	for _, field := range nav[:len(nav)-1] {
		// Get value of field in object
		value, ok := obj[field]
		if !ok {
			return nil, false
		}

		// Value should be a (nested) object, hence update obj with value for next iteration
		obj, ok = value.(map[string]interface{})
		if !ok {
			return nil, false
		}
	}

	// Last field in navigation holds the actual value
	field := nav[len(nav)-1]
	return obj[field], true
}

// parseSecret returns value navigated by given fields on secret object.
// Note that a secret returned from Vault is effectively a JSON object.
func parseSecret(secret *vault.Secret, path string, fields string) ([]byte, error) {
	value, ok := valueOf(secret.Data, fields)
	if !ok {
		err := fmt.Errorf("HashiCorp Vault provider expects secret in '%s' at '%s'", fields, path)
		return nil, err
	}

	// Secret value must be either string or bytes
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		err := fmt.Errorf("HashiCorp Vault provider expects the secret to be a string or byte[], got %T", v)
		return nil, err
	}
}

// GetValue obtains a value by id. The id should contain the path in Vault to the secret. It may be appended with a
// hash following the object property path to the secret value; defaults to DefaultField.
// For example:
//   - `kv/database/password` returns the value of field `value` in the secret object at given path.
//   - `kv/database#password` returns the value of field `password` in the secret object at path `kv/database`.
//   - `secret/data/database#data.value` returns the value of field `value` wrapped in object `data` in secret object
//  	at path `secret/data/database`.
// Secrets in Vault are stored as (JSON) objects in the shape of map[string]interface{}. Both path to the secret and
// fields to the value in the secret must follow Vault API client conventions. Please see documentation of Vault for
// details.
func (p *Provider) GetValue(id string) ([]byte, error) {
	path, fields := parseVaultID(id)
	secret, err := p.Client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		err = fmt.Errorf("HashiCorp Vault provider could not find secret '%s'", path)
		return nil, err
	}

	return parseSecret(secret, path, fields)
}

// GetValues takes in variable ids and returns their resolved values. This method is
// needed to the Provider interface
func (p *Provider) GetValues(ids ...string) (map[string]plugin_v1.ProviderResponse, error) {
	return plugin_v1.GetValues(p, ids...)
}
