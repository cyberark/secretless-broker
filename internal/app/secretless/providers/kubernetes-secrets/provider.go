package kubernetes_secrets

import (
	"fmt"
	"log"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/core/v1"
)

// Provider provides data values from Kubernetes Secrets.
type Provider struct {
	Name   string
	Client *KubeClient
}

// ProviderFactory constructs a Provider. The API client is configured from
// in-cluster environment variables and files.
func ProviderFactory(options plugin_v1.ProviderOptions) plugin_v1.Provider {
	var client *KubeClient
	var err error
	if client, err = NewKubeClient(); err != nil {
		log.Panicf("ERROR: Could not create Kubernetes Secrets provider: %s", err)
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

// parseKubernetesSecretID returns the secret id and field name.
func parseKubernetesSecretID(id string) (string, string) {
	tokens := strings.SplitN(id, "#", 2)

	return tokens[0], tokens[1]
}

// GetValue obtains a value by id. Any secret which is stored in Kubernetes Secrets is recognized.
// The datatype returned by Kubernetes Secrets is map[string][]byte. Therefore this provider needs
// to know which field to return from the map. The field to be returned is specified by appending '#fieldName' to the id argument.
func (p Provider) GetValue(id string) ([]byte, error) {
	id, fieldName := parseKubernetesSecretID(id)

	currentNamespace, err := p.Client.CurrentNamespace()
	if err != nil {
		return nil, err
	}

	secret, err := p.Client.GetSecret(currentNamespace, id);
	if err != nil {
		return nil, err
	}

	value, ok := secret.Data[fieldName];
	if !ok {
		return nil, fmt.Errorf("Kubernetes Secrets provider expects the secret '%s' to contain field '%s'", id, fieldName)
	}

	return value, nil
}

// KubeClient represents Kubernetes client and calculated namespace
type KubeClient struct {
	clientset *kubernetes.Clientset
	clientConfig clientcmd.ClientConfig
}

// NewKubeClient creates new Kubernetes API client
func NewKubeClient() (*KubeClient, error) {
	// creates the in-cluster config
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubeClient{
		clientset: clientset,
		clientConfig: clientConfig,
	}, nil
}

// CurrentNamespace returns the in-cluster current namespace
func (c *KubeClient) CurrentNamespace() (string, error) {
	namespace, _, err := c.clientConfig.Namespace()

	return namespace, err
}

// GetSecret returns Secret in the given namespace with the given name
func (c *KubeClient) GetSecret(namespace, name string) (*v1.Secret, error) {
	return c.clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}
