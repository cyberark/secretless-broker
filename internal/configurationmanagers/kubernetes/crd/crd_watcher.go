package crd

import (
	"context"
	"log"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	secretlessClientset "github.com/cyberark/secretless-broker/pkg/k8sclient/clientset/versioned"
)

// ResourceEventHandler is the interface for handling CRD push notification
type ResourceEventHandler interface {
	CRDAdded(*api_v1.Configuration)
	CRDDeleted(*api_v1.Configuration)
	CRDUpdated(*api_v1.Configuration, *api_v1.Configuration)
}

// RegisterCRDListener registers a CRD push-notification handler to the available
// k8s cluster
func RegisterCRDListener(namespace string, configSpec string, resourceEventHandler ResourceEventHandler) error {
	log.Printf("%s: Registering CRD watcher...", PluginName)

	clientConfig, err := NewKubernetesConfig()
	if err != nil {
		return err
	}

	clientset, err := secretlessClientset.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	// TODO: Watch for CRD availability

	// TODO: We might not want to listen in on all namespace changes
	watchList := &cache.ListWatch{
		ListFunc: func(listOpts meta_v1.ListOptions) (result runtime.Object, err error) {
			return clientset.SecretlessV1().Configurations(namespace).List(context.TODO(), listOpts)
		},
		WatchFunc: func(listOpts meta_v1.ListOptions) (watch.Interface, error) {
			return clientset.SecretlessV1().Configurations(namespace).Watch(context.TODO(), listOpts)
		},
	}

	_, controller := cache.NewInformer(
		watchList,
		&api_v1.Configuration{},
		CRDForcedRefreshInterval,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				configObj := obj.(*api_v1.Configuration)
				if configObj.ObjectMeta.Name != configSpec {
					return
				}

				log.Printf("%s: Add configuration event", PluginName)
				log.Println(configObj.ObjectMeta.Name)
				resourceEventHandler.CRDAdded(configObj)
			},
			DeleteFunc: func(obj interface{}) {
				configObj := obj.(*api_v1.Configuration)
				if configObj.ObjectMeta.Name != configSpec {
					return
				}

				log.Printf("%s: Delete configuration event", PluginName)
				resourceEventHandler.CRDDeleted(configObj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldConfigObj := oldObj.(*api_v1.Configuration)
				if oldConfigObj.ObjectMeta.Name != configSpec {
					return
				}

				log.Printf("%s: Update/refresh configuration event", PluginName)
				newConfigObj := newObj.(*api_v1.Configuration)
				resourceEventHandler.CRDUpdated(oldConfigObj, newConfigObj)
			},
		},
	)

	go controller.Run(wait.NeverStop)

	return nil
}
