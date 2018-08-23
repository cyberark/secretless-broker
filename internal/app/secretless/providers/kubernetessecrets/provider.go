package kubernetessecrets

import (
	"fmt"
	"strings"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// Provider provides data values from Kubernetes Secrets.
type Provider struct {
	Name   string
	Client *KubeClient
}

// ProviderFactory constructs a Provider. The API client is configured from
// in-cluster environment variables and files.
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	var client *KubeClient
	var err error
	if client, err = NewKubeClient(); err != nil {
		return nil, fmt.Errorf("ERROR: Could not create Kubernetes Secrets provider: %s", err)
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

// GetValues resolves multiple Kubernetes secrets into values
func (p *Provider) GetValues(variables []config.Variable) (map[string][]byte, error) {
	result := map[string][]byte{}
	for _, variable := range variables {
		value, err := p.GetValue(variable.ID)
			if err != nil {
				return nil, err
			}
		result[variable.Name] = value
	}
	return result, nil
}

// GetValue obtains a value by id. Any secret which is stored in Kubernetes Secrets is recognized.
// The data type returned by Kubernetes Secrets is map[string][]byte. Therefore this provider needs
// to know which field to return from the map. The field to be returned is specified by appending '#fieldName' to the id argument.
func (p *Provider) GetValue(id string) ([]byte, error) {
	tokens := strings.SplitN(id, "#", 2)

	if len(tokens) != 2 {
		return nil, fmt.Errorf("Kubernetes secret id must contain secret name and field name in the format secretName#fieldName, received '%s'", id)
	}

	secretName, fieldName := tokens[0], tokens[1]
	if fieldName == "" {
		return nil, fmt.Errorf("name of field missing from Kubernetes secret id '%s'", id)
	}

	currentNamespace, err := p.Client.CurrentNamespace()
	if err != nil {
		return nil, err
	}

	secret, err := p.Client.GetSecret(currentNamespace, secretName)
	if err != nil {
		return nil, err
	}

	value, ok := secret.Data[fieldName]
	if !ok {
		return nil, fmt.Errorf("could not find field '%s' in Kubernetes secret '%s'", fieldName, secretName)
	}

	return value, nil
}
