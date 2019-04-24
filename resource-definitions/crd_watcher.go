package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	secretlessClientset "github.com/cyberark/secretless-broker/pkg/k8sclient/clientset/versioned"
)

func getHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return os.Getenv("USERPROFILE")
}

func main() {
	log.Println("Secretless CRD watcher starting up...")

	var kubeConfig *string
	if home := getHomeDir(); home != "" {
		log.Println("Using home dir config...")
		kubeConfig = flag.String("kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		log.Println("Using passed in file config...")
		kubeConfig = flag.String("kubeconfig",
			"",
			"absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Try to use file-based config first
	clientConfig, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		log.Println(err)
	}

	// Otherwise try using in-cluster service account
	if clientConfig == nil {
		log.Println("Fetching cluster config...")
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Create new clientset
	clientset, err := secretlessClientset.NewForConfig(clientConfig)
	if err != nil {
		log.Fatalln(err)
	}

	// List the available configurations
	list, err := clientset.SecretlessV1().Configurations("default").List(meta_v1.ListOptions{})
	log.Printf("Available configs: %v", len(list.Items))
	for _, config := range list.Items {
		yamlContent, err := yaml.Marshal(&config)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Config: \n%v", string(yamlContent))
	}

	watchList := &cache.ListWatch{
		ListFunc: func(listOpts meta_v1.ListOptions) (result runtime.Object, err error) {
			return clientset.SecretlessV1().Configurations(meta_v1.NamespaceAll).List(listOpts)
		},
		WatchFunc: func(listOpts meta_v1.ListOptions) (watch.Interface, error) {
			return clientset.SecretlessV1().Configurations(meta_v1.NamespaceAll).Watch(listOpts)
		},
	}

	// Watch for changes in Example objects and fire Add, Delete, Update callbacks
	log.Println("Watching for changes...")
	_, controller := cache.NewInformer(
		watchList,
		&api_v1.Configuration{},
		10*time.Minute,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Println("Add")
				yamlContent, _ := yaml.Marshal(&obj)
				log.Printf("Add event: \n%v\n", string(yamlContent))
			},
			DeleteFunc: func(obj interface{}) {
				log.Println("Delete")
				yamlContent, _ := yaml.Marshal(&obj)
				log.Printf("Delete event: \n%v\n", string(yamlContent))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Println("Update")
				oldYamlContent, _ := yaml.Marshal(&oldObj)
				newYamlContent, _ := yaml.Marshal(&newObj)
				log.Println("Update event:")
				log.Printf("Old:\n%v\nNew:\n%v\n", string(oldYamlContent),
					string(newYamlContent))
			},
		},
	)

	go controller.Run(wait.NeverStop)

	// Wait forever
	select {}
}
