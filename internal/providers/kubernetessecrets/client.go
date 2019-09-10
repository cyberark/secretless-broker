package kubernetessecrets

import (
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// NewSecretsClient creates a SecretsClient for working with Secrets
// in the pod namespace.
// A SecretsClient uses an underlying Kubernetes API client
// initialised with in-cluster config.
func NewSecretsClient() (typedv1.SecretInterface, error) {
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

	// determine pod namespace
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, err
	}

	return clientset.CoreV1().Secrets(namespace), nil
}
