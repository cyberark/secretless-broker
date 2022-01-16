package conjur

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	authnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

const authenticatorTokenFile = "/run/conjur/access-token"

var supportedAuthentications = []string{"authn-k8s", "authn-jwt"}

// Provider provides data values from the Conjur vault.
type Provider struct {
	// Data related to Provider config
	AuthenticationMutex *sync.Mutex
	Authenticator       authenticator.Authenticator
	AuthenticatorConfig authnConfig.Configuration
	Config              conjurapi.Config
	Conjur              *conjurapi.Client
	Version             string
	Name                string

	// Credentials for API-key based auth
	APIKey   string
	Username string

	// Authn URL for K8s-authenticator based auth
	AuthnURL string
}

// ProviderFactory constructs a Conjur Provider. The API client is configured from
// environment variables.
// To authenticate with Conjur, you can provide Secretless with:
// - A Conjur username and API key
// - A path to a file where a Conjur access token is stored
// - Config info to use the Conjur k8s authenticator client to retrieve an access token
//   from Conjur (i.e. Conjur version, account, authn url, username, and SSL cert)
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("ERROR: Conjur provider could not load configuration: %s", err)
	}

	var apiKey, authnURL, tokenFile, username, version string
	var conjurAuthenticator authenticator.Authenticator
	var conjurAuthenticatorConf authnConfig.Configuration
	var conjur *conjurapi.Client
	var provider *Provider

	authenticationMutex := &sync.Mutex{}

	username = os.Getenv("CONJUR_AUTHN_LOGIN")
	apiKey = os.Getenv("CONJUR_AUTHN_API_KEY")
	tokenFile = os.Getenv("CONJUR_AUTHN_TOKEN_FILE")
	authnURL = os.Getenv("CONJUR_AUTHN_URL")
	version = os.Getenv("CONJUR_VERSION")

	if len(version) == 0 {
		version = "5"
	}

	provider = &Provider{
		Name:                options.Name,
		Config:              config,
		Username:            username,
		AuthenticatorConfig: conjurAuthenticatorConf,
		APIKey:              apiKey,
		AuthnURL:            authnURL,
		AuthenticationMutex: authenticationMutex,
		Version:             version,
	}

	switch {
	case provider.Username != "" && provider.APIKey != "":
		// Conjur provider using API key
		log.Printf("Info: Conjur provider using API key-based authentication")
		if conjur, err = conjurapi.NewClientFromKey(provider.Config, authn.LoginPair{provider.Username, provider.APIKey}); err != nil {
			return nil, fmt.Errorf("ERROR: Could not create new Conjur provider: %s", err)
		}
	case tokenFile != "":
		// Conjur provider using access token
		log.Printf("Info: Conjur provider using access token-based authentication")
		if conjur, err = conjurapi.NewClientFromTokenFile(provider.Config, tokenFile); err != nil {
			return nil, fmt.Errorf("ERROR: Could not create new Conjur provider: %s", err)
		}
	case urlSupported(provider.AuthnURL):
		log.Printf("Info: Conjur provider doing auhtentication to conjur to endpoint %s", provider.AuthnURL)

		// Load the authenticator with the config from the environment, and log in to Conjur
		conjurAuthenticatorConf, err = authnConfig.NewConfigFromEnv()
		if err != nil {
			return nil, err
		}
		conjurAuthenticator, err = authenticator.NewAuthenticator(conjurAuthenticatorConf)

		if err != nil {
			return nil, fmt.Errorf("ERROR: Conjur provider could not retrieve access token using the authenticator client: %s", err)
		}
		provider.Authenticator = conjurAuthenticator
		provider.AuthenticatorConfig = conjurAuthenticatorConf

		refreshErr := provider.fetchAccessToken()
		if refreshErr != nil {
			return nil, refreshErr
		}

		go func() {
			// Sleep until token needs refresh
			time.Sleep(provider.AuthenticatorConfig.GetTokenTimeout())

			// Authenticate in a loop
			err := provider.fetchAccessTokenLoop()

			// On repeated errors in getting the token, we need to exit the
			// broker since the provider cannot be used.
			if err != nil {
				log.Fatal(err)
			}
		}()

		// Once the token file has been loaded, create a new instance of the Conjur client
		if conjur, err = conjurapi.NewClientFromTokenFile(provider.Config, authenticatorTokenFile); err != nil {
			return nil, fmt.Errorf("ERROR: Could not create new Conjur provider: %s", err)
		}
	default:
		return nil, errors.New("ERROR: Unable to construct a Conjur provider client from the available credentials")
	}

	provider.Conjur = conjur

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

// GetValue obtains a value by ID. The recognized IDs are:
//	* "accessToken"
// 	* Any Conjur variable ID
func (p *Provider) GetValue(id string) ([]byte, error) {
	var err error

	if id == "accessToken" {
		if p.Username != "" && p.APIKey != "" {
			// TODO: Use a cached access token from the client, once it's exposed
			return p.Conjur.Authenticate(authn.LoginPair{
				p.Username,
				p.APIKey,
			})
		}
		return nil, errors.New("Error: Conjur provider can't provide an accessToken unless username and apiKey credentials are provided")
	}

	// If using the Conjur Kubernetes authenticator, ensure that the
	// Conjur API is using the current access token
	if p.AuthnURL != "" {
		if p.Conjur, err = conjurapi.NewClientFromTokenFile(p.Config, authenticatorTokenFile); err != nil {
			log.Fatalf("ERROR: Could not create new Conjur provider: %s", err)
		}
	}

	tokens := strings.SplitN(id, ":", 3)
	switch len(tokens) {
	case 1:
		tokens = []string{p.Config.Account, "variable", tokens[0]}
	case 2:
		tokens = []string{p.Config.Account, tokens[0], tokens[1]}
	}

	return p.Conjur.RetrieveSecret(strings.Join(tokens, ":"))
}

// fetchAccessToken uses the Conjur Kubernetes authenticator
// to authenticate with Conjur and retrieve a new time-limited
// access token.
// fetchAccessToken carries out retry with exponential backoff
func (p *Provider) fetchAccessToken() error {
	// Configure exponential backoff
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second
	expBackoff.RandomizationFactor = 0.5
	expBackoff.Multiplier = 2
	expBackoff.MaxInterval = 15 * time.Second
	expBackoff.MaxElapsedTime = 2 * time.Minute

	// Authenticate with retries on failure with exponential backoff
	err := backoff.Retry(func() error {
		// Lock the authenticatorMutex
		p.AuthenticationMutex.Lock()
		defer p.AuthenticationMutex.Unlock()

		log.Printf("Info: Conjur provider is authenticating ...")
		if err := p.Authenticator.Authenticate(); err != nil {
			log.Printf("Info: Conjur provider received an error on authenticate: %s", err.Error())
			return err
		}

		return nil
	}, expBackoff)

	if err != nil {
		return fmt.Errorf("Error: Conjur provider unable to authenticate; backoff exhausted: %s", err.Error())
	}

	return nil
}

// fetchAccessTokenLoop runs authenticate in an infinite loop
// punctuated by by sleeps of duration TokenRefreshTimeout
func (p *Provider) fetchAccessTokenLoop() error {
	if p.Authenticator == nil {
		return errors.New("Error: Conjur Kubernetes authenticator must be instantiated before access token may be refreshed")
	}

	// Fetch the access token in a loop
	for {
		err := p.fetchAccessToken()
		if err != nil {
			return err
		}

		// sleep until token needs refresh
		time.Sleep(p.AuthenticatorConfig.GetTokenTimeout())
	}
}

func urlSupported(url string) bool {
	if url == "" {
		return false
	}

	for _, authnStrategy := range supportedAuthentications {
		if strings.Contains(url, authnStrategy) {
			return true
		}
	}

	return false
}
