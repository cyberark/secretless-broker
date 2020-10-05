package awssecrets

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

// Provider provides data values from AWS Secrets Manager.
type Provider struct {
	Name   string
	Client *secretsmanager.SecretsManager
}

// ProviderFactory constructs a Provider. The API client is configured from
// in-cluster environment variables and files.
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {

	// All clients require a Session. The Session provides the client with
	// shared configuration such as region, endpoint, and credentials. A
	// Session should be shared where possible to take advantage of
	// configuration and credential caching.
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("ERROR: Could not create AWS Secrets provider: %s", err)
	}
	// Create a new instance of the service's client with a Session.
	client := secretsmanager.New(sess)

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

	req, resp := client.GetSecretValueRequest(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(id),
	})

	err := req.Send()
	if err != nil { // resp is now filled
		return nil, err
	}

	if resp.SecretString != nil {
		return []byte(*resp.SecretString), nil
	}

	return resp.SecretBinary, nil
}
