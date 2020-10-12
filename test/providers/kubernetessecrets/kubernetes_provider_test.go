package main

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
	"github.com/cyberark/secretless-broker/internal/providers/kubernetessecrets"
)

func TestKubernetes_Provider(t *testing.T) {
	var (
		err                error
		provider           plugin_v1.Provider
		kubernetesProvider *kubernetessecrets.Provider
	)

	var testSecretsClient = testclient.NewSimpleClientset().CoreV1().Secrets("some-namespace")

	_, err = testSecretsClient.Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "database",
		},
		Data: map[string][]byte{
			"password": []byte("secret"),
		},
	})
	if err != nil {
		panic(fmt.Errorf("unable to create secret on test client: %s", err))
	}

	expectedName := "kubernetes"

	options := plugin_v1.ProviderOptions{
		Name: expectedName,
	}

	Convey("Can create the Kubernetes provider", t, func() {
		provider, err = providers.ProviderFactories[expectedName](options)
		So(err, ShouldBeNil)

		var ok bool
		kubernetesProvider, ok = provider.(*kubernetessecrets.Provider)
		So(ok, ShouldBeTrue)

		kubernetesProvider.SecretsClient = testSecretsClient
	})

	Convey("Has the expected provider name", t, func() {
		So(provider.GetName(), ShouldEqual, expectedName)
	})

	Convey("Can provide a secret", t, func() {
		id := "database#password"
		values, err := provider.GetValues(id)
		So(err, ShouldBeNil)
		So(values, ShouldContainKey, id)
		So(values[id].Error, ShouldBeNil)
		So(string(values[id].Value), ShouldEqual, "secret")
	})

	Convey("Reports", t, func() {
		for _, testCase := range reportsTestCases {
			Convey(
				testCase.description,
				reports(provider, testCase.id, testCase.expectedErrString),
			)
		}
	})
}

type reportsTestCase struct {
	description       string
	id                string
	expectedErrString string
}

func reports(provider plugin_v1.Provider, id string, expectedErrString string) func() {
	return func() {
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values, ShouldContainKey, id)
		So(values[id].Value, ShouldBeNil)
		So(values[id].Error, ShouldNotBeNil)
		So(values[id].Error.Error(), ShouldEqual, expectedErrString)
	}
}

var reportsTestCases = []reportsTestCase{
	{
		description:       "Reports when the secret id does not contain a field name",
		id:                "foobar",
		expectedErrString: "Kubernetes secret id must contain secret name and field name in the format secretName#fieldName, received 'foobar'",
	},
	{
		description:       "Reports when the secret id has empty field name",
		id:                "foobar#",
		expectedErrString: "field name missing from Kubernetes secret id 'foobar#'",
	},
	{
		description:       "Reports when Kubernetes is unable to find secret",
		id:                "foobar#maybe",
		expectedErrString: "could not find Kubernetes secret from 'foobar#maybe'",
	},
	{
		description:       "Reports when Kubernetes is unable to find field name in secret",
		id:                "database#missing",
		expectedErrString: "could not find field 'missing' in Kubernetes secret 'database'",
	},
}
