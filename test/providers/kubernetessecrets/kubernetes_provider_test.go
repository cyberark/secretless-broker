package main

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/testutils"
	"github.com/cyberark/secretless-broker/internal/providers"
	"github.com/cyberark/secretless-broker/internal/providers/kubernetessecrets"
)

var mockSecrets = map[string]map[string][]byte{
	"database": {
		"password": []byte("secret-value"),
	},
	"server1": {
		"api-key": []byte("api-key-value"),
		"token":   []byte("token-value"),
	},
}

func TestKubernetes_Provider(t *testing.T) {
	var (
		err                error
		provider           plugin_v1.Provider
		kubernetesProvider *kubernetessecrets.Provider
	)

	var testSecretsClient = testclient.NewSimpleClientset().CoreV1().Secrets(
		"some-namespace",
	)

	for name, data := range mockSecrets {
		_, err = testSecretsClient.Create(
			context.TODO(),
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: data,
			},
			metav1.CreateOptions{},
		)
		if err != nil {
			panic(fmt.Errorf("unable to create secret on test client: %s", err))
		}
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
		So(string(values[id].Value), ShouldEqual, "secret-value")
	})

	Convey("Reports", t, func() {
		for _, testCase := range reportsTestCases {
			Convey(
				testCase.Description,
				testutils.Reports(provider, testCase.ID, testCase.ExpectedErrString),
			)
		}
	})

	Convey(
		"Multiple Provides ",
		t,
		testutils.CanProvideMultiple(
			provider,
			map[string]string{
				"database#password": "secret-value",
				"server1#api-key":   "api-key-value",
				"server1#token":     "token-value",
			},
		),
	)
}

var reportsTestCases = []testutils.ReportsTestCase{
	{
		Description: "Reports when the secret id does not contain a field name",
		ID:          "foobar",
		ExpectedErrString: "Kubernetes secret id must contain secret name and field name " +
			"in the format secretName#fieldName, received 'foobar'",
	},
	{
		Description:       "Reports when the secret id has empty field name",
		ID:                "foobar#",
		ExpectedErrString: "field name missing from Kubernetes secret id 'foobar#'",
	},
	{
		Description:       "Reports when Kubernetes is unable to find secret",
		ID:                "foobar#maybe",
		ExpectedErrString: "could not find Kubernetes secret from 'foobar#maybe'",
	},
	{
		Description:       "Reports when Kubernetes is unable to find field name in secret",
		ID:                "database#missing",
		ExpectedErrString: "could not find field 'missing' in Kubernetes secret 'database'",
	},
}
