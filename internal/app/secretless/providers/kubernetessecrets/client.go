package kubernetessecrets

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewKubeClient creates a Kubernetes API client using in-cluster config
func NewKubeClient() (kubernetes.Interface, error) {
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

	return clientset, nil
}

// GetSecret fetches a Secret with a given name
func GetSecret(kc kubernetes.Interface, name string) (*v1.Secret, error) {
	return kc.CoreV1().Secrets("").Get(name, metav1.GetOptions{})
}
