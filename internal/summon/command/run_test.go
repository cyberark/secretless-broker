package command

import (
	"testing"

	"github.com/cyberark/summon/secretsyml"
	"github.com/stretchr/testify/assert"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

func Test_resolveSecrets(t *testing.T) {
	p := plugin_v1.MockProvider{}
	secretsMap := func() secretsyml.SecretsMap {
		return secretsyml.SecretsMap{
			"bar": secretsyml.SecretSpec{
				Tags: []secretsyml.YamlTag{secretsyml.Literal},
				Path: "bar_path",
			},
			"baz": secretsyml.SecretSpec{
				Tags: []secretsyml.YamlTag{secretsyml.Var},
				Path: "baz_path",
			},
			"foo": secretsyml.SecretSpec{
				Tags: []secretsyml.YamlTag{secretsyml.Var, secretsyml.File},
				Path: "foo_path",
			},
		}
	}

	t.Run("Resolves variable and literal secrets", func(t *testing.T) {
		secrets, err := resolveSecrets(
			&p,
			secretsMap(),
		)

		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(
			t,
			secrets,
			map[string]string{
				"bar": "bar_path",
				"baz": "baz_path_value",
				"foo": "foo_path_value",
			})
	})

	t.Run("Reports errors", func(t *testing.T) {
		sm := secretsMap()
		sm["err1"] = secretsyml.SecretSpec{
			Tags: []secretsyml.YamlTag{secretsyml.Var, secretsyml.File},
			Path: "err_first",
		}
		sm["err2"] = secretsyml.SecretSpec{
			Tags: []secretsyml.YamlTag{secretsyml.Var},
			Path: "err_second",
		}

		secrets, err := resolveSecrets(
			&p,
			sm,
		)

		if !assert.EqualError(t, err, "err_first_value\nerr_second_value") {
			return
		}
		assert.Nil(t, secrets)
	})
}

func Test_buildEnvironment(t *testing.T) {
	tempFactory := NewTempFactory("")

	defer tempFactory.Cleanup()

	env, err := buildEnvironment(
		map[string]string{
			"bar": "bar_val",
			"baz": "baz_val",
			"foo": "foo_val",
		},
		secretsyml.SecretsMap{
			"bar": secretsyml.SecretSpec{
				Tags: []secretsyml.YamlTag{secretsyml.Literal, secretsyml.File},
			},
			"baz": secretsyml.SecretSpec{
				Tags: []secretsyml.YamlTag{secretsyml.Literal, secretsyml.File},
			},
			"foo": secretsyml.SecretSpec{
				Tags: []secretsyml.YamlTag{secretsyml.Literal},
			},
		},
		&tempFactory,
	)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(
		t,
		env,
		[]string{
			"bar=" + tempFactory.files[0],
			"baz=" + tempFactory.files[1],
			"foo=foo_val",
		},
	)
}
