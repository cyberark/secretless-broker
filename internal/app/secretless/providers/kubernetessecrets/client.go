package kubernetessecrets

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubeClient represents a Kubernetes client
type KubeClient struct {
	Client    kubernetes.Interface
}

// NewKubeClient creates new Kubernetes API client
func NewKubeClient() (*KubeClient, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubeClient{
		Client: clientset,
	}, nil
}

// GetSecret fetches a Secret with a given name
func (c *KubeClient) GetSecret(name string) (*v1.Secret, error) {
	return c.Client.CoreV1().Secrets("").Get(name, metav1.GetOptions{})
}
