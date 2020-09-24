package main

import (
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

	testSecretsClient.Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "database",
		},
		Data: map[string][]byte{
			"password": []byte("secret"),
		},
	})

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

	Convey("Reports when the secret id does not contain a field name", t, func() {
		values, err := provider.GetValues("foobar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Kubernetes secret id must contain secret name and field name in the format secretName#fieldName, received 'foobar'")
		So(values, ShouldBeNil)
	})

	Convey("Reports when the secret id has empty field name", t, func() {
		values, err := provider.GetValues("foobar#")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "field name missing from Kubernetes secret id 'foobar#'")
		So(values, ShouldBeNil)
	})

	Convey("Reports when Kubernetes is unable to find secret", t, func() {
		values, err := provider.GetValues("foobar#maybe")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "could not find Kubernetes secret from 'foobar#maybe'")
		So(values, ShouldBeNil)
	})

	Convey("Reports when Kubernetes is unable to find field name in secret", t, func() {
		values, err := provider.GetValues("database#missing")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "could not find field 'missing' in Kubernetes secret 'database'")
		So(values, ShouldBeNil)
	})

	Convey("Can provide a secret", t, func() {
		values, err := provider.GetValues("database#password")
		So(err, ShouldBeNil)
		So(string(values[0]), ShouldEqual, "secret")
	})
}
