package kubernetessecrets

import (
	"fmt"
	"strings"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

// Provider provides data values from Kubernetes Secrets.
type Provider struct {
	Name   string
	Client kubernetes.Interface
}

// ProviderFactory constructs a Provider. The API client is configured from
// in-cluster environment variables and files.
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	client, err := NewKubeClient()

	if err != nil {
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
		return nil, fmt.Errorf("field name missing from Kubernetes secret id '%s'", id)
	}

	secret, err := GetSecret(p.Client, secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("could not find Kubernetes secret from '%s'", id)
		}
		return nil, err
	}

	value, ok := secret.Data[fieldName]
	if !ok {
		return nil, fmt.Errorf("could not find field '%s' in Kubernetes secret '%s'", fieldName, secretName)
	}

	return value, nil
}
