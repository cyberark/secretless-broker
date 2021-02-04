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

// Provider provides data values from the Conjur vault.
type Provider struct {
	// Data related to Provider config
	AuthenticationMutex *sync.Mutex
	Authenticator       *authenticator.Authenticator
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

func hasField(field string, params *map[string]string) (ok bool) {
	_, ok = (*params)[field]
	return
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
	var authenticator *authenticator.Authenticator
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
		APIKey:              apiKey,
		AuthnURL:            authnURL,
		AuthenticationMutex: authenticationMutex,
		Version:             version,
	}

	if provider.Username != "" && provider.APIKey != "" {
		log.Printf("Info: Conjur provider using API key-based authentication")
		if conjur, err = conjurapi.NewClientFromKey(provider.Config, authn.LoginPair{provider.Username, provider.APIKey}); err != nil {
			return nil, fmt.Errorf("ERROR: Could not create new Conjur provider: %s", err)
		}
	} else if tokenFile != "" {
		log.Printf("Info: Conjur provider using access token-based authentication")
		if conjur, err = conjurapi.NewClientFromTokenFile(provider.Config, tokenFile); err != nil {
			return nil, fmt.Errorf("ERROR: Could not create new Conjur provider: %s", err)
		}
	} else if provider.AuthnURL != "" && strings.Contains(provider.AuthnURL, "authn-k8s") {
		log.Printf("Info: Conjur provider using Kubernetes authenticator-based authentication")

		// Load the authenticator with the config from the environment, and log in to Conjur
		if authenticator, err = loadAuthenticator(provider.AuthnURL, provider.Version, provider.Config); err != nil {
			return nil, fmt.Errorf("ERROR: Conjur provider could not retrieve access token using the authenticator client: %s", err)
		}
		provider.Authenticator = authenticator

		refreshErr := provider.fetchAccessToken()
		if refreshErr != nil {
			return nil, refreshErr
		}

		go func() {
			// Sleep until token needs refresh
			time.Sleep(provider.Authenticator.Config.TokenRefreshTimeout)

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
	} else {
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

// loadAuthenticator returns a Conjur Kubernetes authenticator client
// that has performed the login process to retrieve the signed certificate
// from Conjur
// The authenticator will be used to retrieve a time-limited access token
// This method requires CONJUR_ACCOUNT, CONJUR_AUTHN_URL, CONJUR_AUTHN_LOGIN, and
// CONJUR_SSL_CERTIFICATE/CONJUR_CERT_FILE env vars to be present
// if CONJUR_VERSION is not present, it defaults to "5"
// Currently the deployment manifest for Secretless must also specify
// MY_POD_NAMESPACE and MY_POD_NAME from the pod metadata, but there is a GH
// issue logged in the authenticator for doing this via the Kubernetes API
func loadAuthenticator(authnURL string, version string,
	providerConfig conjurapi.Config) (*authenticator.Authenticator, error) {

	var err error

	// Check that required environment variables are set
	config, err := authnConfig.NewFromEnv()
	if err != nil {
		return nil, err
	}

	// Create new Authenticator
	authenticator, err := authenticator.New(*config)
	if err != nil {
		return nil, err
	}

	return authenticator, nil
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

		log.Printf("Info: Conjur provider is authenticating as %s ...", p.Authenticator.Config.Username)
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
		time.Sleep(p.Authenticator.Config.TokenRefreshTimeout)
	}
}
