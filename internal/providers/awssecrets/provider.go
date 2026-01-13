package awssecrets

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

// Provider provides data values from AWS Secrets Manager.
type Provider struct {
	Name   string
	Client *secretsmanager.Client
}

// ProviderFactory constructs a Provider. The API client is configured from
// in-cluster environment variables and files.
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {

	// v2: load default config (honors AWS_REGION/AWS_DEFAULT_REGION, shared config & creds)
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return nil, fmt.Errorf("ERROR: Could not create AWS Secrets provider: %s", err)
	}

	// v2: build service client from config
	client := secretsmanager.NewFromConfig(cfg)

	provider := &Provider{
		Name:   options.Name,
		Client: client,
	}

	return provider, nil
}

// GetName returns the name of the provider
func (p *Provider) GetName() string {
	return p.Name
}

// GetValues takes in variable ids and returns their resolved values. This method is
// needed to the Provider interface
func (p *Provider) GetValues(ids ...string) (map[string]plugin_v1.ProviderResponse, error) {
	return plugin_v1.GetValues(p, ids...)
}

// GetValue obtains a secret value by id.
func (p *Provider) GetValue(id string) ([]byte, error) {
	client := p.Client

	out, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(id),
	})
	if err != nil {
		return nil, err
	}
	if out.SecretString != nil {
		return []byte(*out.SecretString), nil
	}
	return out.SecretBinary, nil
}
