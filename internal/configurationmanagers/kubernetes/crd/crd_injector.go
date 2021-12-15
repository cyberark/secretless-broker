package crd

import (
	"context"
	"fmt"
	"log"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This schema should be kept up-to-date to match secretless-resource-definition.yaml
// and the Config struct in pkg/secretless/config/v1/config.go
var secretlessCRD = &apiextensionsv1.CustomResourceDefinition{
	ObjectMeta: meta_v1.ObjectMeta{
		Name: CRDFQDNName,
	},
	Spec: apiextensionsv1.CustomResourceDefinitionSpec{
		Group: CRDGroupName,
		Names: apiextensionsv1.CustomResourceDefinitionNames{
			Kind:       strings.Title(CRDLongName),
			Plural:     CRDName,
			ShortNames: CRDShortNames,
		},
		Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
			{
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

func createCRD(apiExtClient *apiextensionsclientset.Clientset) error {
	res, err := apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Create(
		context.TODO(), secretlessCRD, meta_v1.CreateOptions{})

	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("%s: ERROR: Could not create Secretless CRD: %v - %v", PluginName,
			err, res)
	}

	if !apierrors.IsAlreadyExists(err) {
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
