package crd

import (
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return os.Getenv("USERPROFILE")
}

// NewKubernetesConfig tries to guess the current k8s configuration for the CRD
// plugin to use
func NewKubernetesConfig() (config *rest.Config, err error) {
	// Try to use file-based config first
	var kubeConfig string
	if home := getHomeDir(); home != "" {
		log.Printf("%s: Using home dir config...", PluginName)

		kubeConfig = filepath.Join(home, ".kube", "config")
		if _, ok := os.Stat(kubeConfig); ok == nil {
			if config, err = clientcmd.BuildConfigFromFlags("", kubeConfig); err != nil {
				return
			}
		} else {
			log.Printf("%s: Skipping home dir config since %s does not exist", PluginName, kubeConfig)
		}
	}

	// Otherwise try using in-cluster service account
	if config == nil {
		log.Printf("%s: Fetching cluster config...", PluginName)
		if config, err = rest.InClusterConfig(); err != nil {
			return
		}
	}

	return config, err
}
