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

	Convey("Reports when the secret id does not contain a field name", t, func() {
		values, err := provider.GetValues("foobar")
		So(err, ShouldBeNil)
		So(values["foobar"], ShouldNotBeNil)
		So(values["foobar"].Value, ShouldBeNil)
		So(values["foobar"].Error, ShouldNotBeNil)
		So(values["foobar"].Error.Error(), ShouldEqual, "Kubernetes secret id must contain secret name and field name in the format secretName#fieldName, received 'foobar'")
	})

	Convey("Reports when the secret id has empty field name", t, func() {
		values, err := provider.GetValues("foobar#")
		So(err, ShouldBeNil)
		So(values["foobar#"], ShouldNotBeNil)
		So(values["foobar#"].Value, ShouldBeNil)
		So(values["foobar#"].Error, ShouldNotBeNil)
		So(values["foobar#"].Error.Error(), ShouldEqual, "field name missing from Kubernetes secret id 'foobar#'")
	})

	Convey("Reports when Kubernetes is unable to find secret", t, func() {
		values, err := provider.GetValues("foobar#maybe")
		So(err, ShouldBeNil)
		So(values["foobar#maybe"], ShouldNotBeNil)
		So(values["foobar#maybe"].Value, ShouldBeNil)
		So(values["foobar#maybe"].Error, ShouldNotBeNil)
		So(values["foobar#maybe"].Error.Error(), ShouldEqual, "could not find Kubernetes secret from 'foobar#maybe'")
	})

	Convey("Reports when Kubernetes is unable to find field name in secret", t, func() {
		values, err := provider.GetValues("database#missing")
		So(err, ShouldBeNil)
		So(values["database#missing"], ShouldNotBeNil)
		So(values["database#missing"].Value, ShouldBeNil)
		So(values["database#missing"].Error, ShouldNotBeNil)
		So(values["database#missing"].Error.Error(), ShouldEqual, "could not find field 'missing' in Kubernetes secret 'database'")
	})

	Convey("Can provide a secret", t, func() {
		values, err := provider.GetValues("database#password")
		So(err, ShouldBeNil)
		So(values["database#password"], ShouldNotBeNil)
		So(string(values["database#password"].Value), ShouldEqual, "secret")
	})
}
