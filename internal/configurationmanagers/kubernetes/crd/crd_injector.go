package crd

import (
	"context"
	"fmt"
	"log"
	"strings"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createCRD(apiExtClient *apiextensionsclientset.Clientset) error {
	secretlessCRD := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: CRDFQDNName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group: CRDGroupName,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:       strings.Title(CRDLongName),
				Plural:     CRDName,
				ShortNames: CRDShortNames,
			},
			Version: "v1",
			Versions: []apiextensionsv1beta1.CustomResourceDefinitionVersion{
				apiextensionsv1beta1.CustomResourceDefinitionVersion{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Scope: apiextensionsv1beta1.NamespaceScoped,
		},
	}

	res, err := apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(), secretlessCRD, meta_v1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("%s: ERROR: Could not create Secretless CRD: %v - %v", PluginName,
			err, res)
	}

	if apierrors.IsAlreadyExists(err) == false {
		log.Printf("%s: CRD was uccessfully added!", PluginName)
	}

	return nil
}

// InjectCRD adds our CRD to K8s if it is missing
func InjectCRD() error {
	config, err := NewKubernetesConfig()
	if err != nil {
		return err
	}

	apiExtClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return err
	}

	return createCRD(apiExtClient)
}
