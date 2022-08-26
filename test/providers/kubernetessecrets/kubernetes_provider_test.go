package main

import (
	"context"
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/testutils"
	"github.com/cyberark/secretless-broker/internal/providers"
	"github.com/cyberark/secretless-broker/internal/providers/kubernetessecrets"
	"github.com/stretchr/testify/assert"
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

	t.Run("Can create the Kubernetes provider", func(t *testing.T) {
		provider, err = providers.ProviderFactories[expectedName](options)
		assert.NoError(t, err)

		var ok bool
		kubernetesProvider, ok = provider.(*kubernetessecrets.Provider)
		assert.True(t, ok)

		kubernetesProvider.SecretsClient = testSecretsClient
	})

	t.Run("Has the expected provider name", func(t *testing.T) {
		assert.Equal(t, expectedName, provider.GetName())
	})

	t.Run("Can provide a secret", func(t *testing.T) {
		id := "database#password"
		values, err := provider.GetValues(id)
		assert.NoError(t, err)
		assert.Contains(t, values, id)
		assert.NoError(t, values[id].Error)
		assert.Equal(t, "secret-value", string(values[id].Value))
	})

	t.Run("Reports", func(t *testing.T) {
		for _, testCase := range reportsTestCases {
			t.Run(
				testCase.Description,
				testutils.Reports(provider, testCase.ID, testCase.ExpectedErrString),
			)
		}
	})

	t.Run(
		"Multiple Provides ",
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
