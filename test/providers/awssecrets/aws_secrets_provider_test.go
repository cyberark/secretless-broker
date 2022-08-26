package main

import (
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
)

func TestAWSSecrets_Provider(t *testing.T) {
	var err error
	var provider plugin_v1.Provider

	name := "aws"

	options := plugin_v1.ProviderOptions{
		Name: name,
	}

	t.Run("Can create the AWS Secrets provider", func(t *testing.T) {
		provider, err = providers.ProviderFactories[name](options)
		assert.NoError(t, err)
	})

	t.Run("Has the expected provider name", func(t *testing.T) {
		assert.Equal(t, "aws", provider.GetName())
	})
}
