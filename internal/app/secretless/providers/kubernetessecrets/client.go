package kubernetessecrets

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeClient represents Kubernetes client and calculated namespace
type KubeClient struct {
	clientset    *kubernetes.Clientset
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
		clientset:    clientset,
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
