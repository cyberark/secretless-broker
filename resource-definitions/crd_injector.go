package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	GroupName = "secretless" + os.Getenv("SECRETLESS_CRD_SUFFIX") + ".io"
	CRDName   = "configurations." + GroupName
)

// This schema should be updated to match secretless-resource-definition.yaml
// and the Config struct in pkg/secretless/config/v1/config.go
var secretlessCRD = &apiextensionsv1.CustomResourceDefinition{
	ObjectMeta: meta_v1.ObjectMeta{
		Name: CRDName,
	},
	Spec: apiextensionsv1.CustomResourceDefinitionSpec{
		Group: GroupName,
		Names: apiextensionsv1.CustomResourceDefinitionNames{
			Kind:   "Configuration",
			Plural: "configurations",
			ShortNames: []string{
				"sbconfig",
			},
		},
		Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
			apiextensionsv1.CustomResourceDefinitionVersion{
				Name:    "v1",
				Served:  true,
				Storage: true,
				Schema: &apiextensionsv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
						Type: "object",
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"spec": {
								Type: "object",
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"listeners": {
										Type: "array",
										Items: &apiextensionsv1.JSONSchemaPropsOrArray{
											Schema: &apiextensionsv1.JSONSchemaProps{
												Type: "object",
												Properties: map[string]apiextensionsv1.JSONSchemaProps{
													"name": {
														Type: "string",
													},
													"protocol": {
														Type: "string",
													},
													"socket": {
														Type: "string",
													},
													"address": {
														Type: "string",
													},
													"debug": {
														Type: "boolean",
													},
													"caCertFiles": {
														Type: "array",
														Items: &apiextensionsv1.JSONSchemaPropsOrArray{
															Schema: &apiextensionsv1.JSONSchemaProps{
																Type: "string",
															},
														},
													},
												},
											},
										},
									},
									"handlers": {
										Type: "array",
										Items: &apiextensionsv1.JSONSchemaPropsOrArray{
											Schema: &apiextensionsv1.JSONSchemaProps{
												Type: "object",
												Properties: map[string]apiextensionsv1.JSONSchemaProps{
													"name": {
														Type: "string",
													},
													"type": {
														Type: "string",
													},
													"listener": {
														Type: "string",
													},
													"debug": {
														Type: "boolean",
													},
													"match": {
														Type: "array",
														Items: &apiextensionsv1.JSONSchemaPropsOrArray{
															Schema: &apiextensionsv1.JSONSchemaProps{
																Type: "string",
															},
														},
													},
													"credentials": {
														Type: "array",
														Items: &apiextensionsv1.JSONSchemaPropsOrArray{
															Schema: &apiextensionsv1.JSONSchemaProps{
																Type: "object",
																Properties: map[string]apiextensionsv1.JSONSchemaProps{
																	"name": {
																		Type: "string",
																	},
																	"provider": {
																		Type: "string",
																	},
																	"id": {
																		Type: "string",
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Scope: apiextensionsv1.NamespaceScoped,
	},
}

func getHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return os.Getenv("USERPROFILE")
}

func createCRD(apiExtClient *apiextensionsclientset.Clientset) {
	log.Println("Creating CRD...")
	res, err := apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Create(
		context.Background(), secretlessCRD, meta_v1.CreateOptions{})

	if err != nil && !apierrors.IsAlreadyExists(err) {
		log.Fatalf("ERROR: Could not create Secretless CRD: %v - %v", err, res)
	}

	if apierrors.IsAlreadyExists(err) {
		log.Println("ERROR: CRD was already present!")
	} else {
		log.Println("CRD was uccessfully added!")
	}
}

// TODO: Use this to wait for the resources to be available
func waitForCRDAvailability(client *rest.RESTClient) error {
	checkCRDAvailableFunc := func() (bool, error) {
		_, err := client.Get().Resource(CRDName).DoRaw(context.Background())
		if err == nil {
			return true, nil
		}

		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	log.Println("Waiting for CRD to be available...")
	return wait.Poll(200*time.Millisecond, 60*time.Second, checkCRDAvailableFunc)
}

func main() {
	log.Println("Secretless CRD injector starting up...")

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
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		log.Println(err)
	}

	// Otherwise try using in-cluster service account
	if config == nil {
		log.Println("Fetching cluster config...")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Println("Creating K8s client...")
	apiExtClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	createCRD(apiExtClient)
	// waitForCRDAvailability(apiExtClient)

	log.Println("Done!")
}
