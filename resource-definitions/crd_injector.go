package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

func getHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return os.Getenv("USERPROFILE")
}

func createCRD(apiExtClient *apiextensionsclientset.Clientset) {
	secretlessCRD := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: CRDName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group: GroupName,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:   "Configuration",
				Plural: "configurations",
				ShortNames: []string{
					"sbconfig",
				},
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

	log.Println("Creating CRD...")
	res, err := apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(secretlessCRD)

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
		_, err := client.Get().Resource(CRDName).DoRaw()
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
