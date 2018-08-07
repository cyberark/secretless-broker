package conjur

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"

	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
)

// authenticateCycleDuration is the default time the system waits to
// reauthenticate on error when using the authenticator client
const authenticateCycleDuration = 6 * time.Minute
const authenticatorTokenFile = "/run/conjur/access-token"

// Provider provides data values from the Conjur vault.
type Provider struct {
	// Data related to Provider config
	AuthenticationMutex *sync.Mutex
	Authenticator       *authenticator.Authenticator
	Config              conjurapi.Config
	Conjur              *conjurapi.Client
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
func ProviderFactory(options plugin_v1.ProviderOptions) plugin_v1.Provider {
	config := conjurapi.LoadConfig()

	var apiKey, authnURL, tokenFile, username string
	var authenticator *authenticator.Authenticator
	var conjur *conjurapi.Client
	var err error
	var provider Provider

	authenticationMutex := &sync.Mutex{}

	username = os.Getenv("CONJUR_AUTHN_LOGIN")
	apiKey = os.Getenv("CONJUR_AUTHN_API_KEY")
	tokenFile = os.Getenv("CONJUR_AUTHN_TOKEN_FILE")
	authnURL = os.Getenv("CONJUR_AUTHN_URL")

	provider = Provider{
		Name:                options.Name,
		Config:              config,
		Username:            username,
		APIKey:              apiKey,
		AuthnURL:            authnURL,
		AuthenticationMutex: authenticationMutex,
	}

	if provider.Username != "" && provider.APIKey != "" {
		log.Printf("Info: Conjur provider using API key-based authentication")
		if conjur, err = conjurapi.NewClientFromKey(provider.Config, authn.LoginPair{provider.Username, provider.APIKey}); err != nil {
			log.Fatalf("ERROR: Could not create new Conjur provider: %s", err)
		}
	} else if tokenFile != "" {
		log.Printf("Info: Conjur provider using access token-based authentication")
		if conjur, err = conjurapi.NewClientFromTokenFile(provider.Config, tokenFile); err != nil {
			log.Fatalf("ERROR: Could not create new Conjur provider: %s", err)
		}
	} else if provider.AuthnURL != "" && strings.Contains(provider.AuthnURL, "authn-k8s") {
		log.Printf("Info: Conjur provider using Kubernetes authenticator-based authentication")

		// Load the authenticator with the config from the environment, and log in to Conjur
		if authenticator, err = loadAuthenticator(provider.AuthnURL, authenticatorTokenFile, provider.Config); err != nil {
			log.Fatalf("ERROR: Conjur provider could not retrieve access token using the authenticator client: %s", err)
		}
		provider.Authenticator = authenticator

		// Kick off the goroutine that will maintain the Conjur access token
		go provider.refreshAccessToken()

		// Once the token file has been loaded, create a new instance of the Conjur client
		if conjur, err = conjurapi.NewClientFromTokenFile(provider.Config, authenticatorTokenFile); err != nil {
			log.Fatalf("ERROR: Could not create new Conjur provider: %s", err)
		}
	} else {
		log.Fatalln("ERROR: Unable to construct a Conjur provider client from the available credentials")
	}

	provider.Conjur = conjur

	return &provider
}

// GetName returns the name of the provider
func (p Provider) GetName() string {
	return p.Name
}

// GetValue obtains a value by ID. The recognized IDs are:
//	* "accessToken"
// 	* Any Conjur variable ID
func (p Provider) GetValue(id string) ([]byte, error) {
	var err error

	if id == "accessToken" {
		if p.Username != "" && p.APIKey != "" {
			// TODO: Use a cached access token from the client, once it's exposed
			return p.Conjur.Authenticate(authn.LoginPair{
				p.Username,
				p.APIKey,
			})
		}
		return nil, fmt.Errorf("Error: Conjur provider can't provide an accessToken unless username and apiKey credentials are provided")
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
func loadAuthenticator(authnURL string, tokenFile string, config conjurapi.Config) (*authenticator.Authenticator, error) {
	var err error
	var conjurCACert []byte

	// Set the client cert / token paths
	clientCertPath := "/etc/conjur/ssl/client.pem"

	// Check that required environment variables are set
	for _, envvar := range []string{
		"CONJUR_ACCOUNT",
		"CONJUR_AUTHN_LOGIN",
		"MY_POD_NAMESPACE",
		"MY_POD_NAME",
	} {
		if os.Getenv(envvar) == "" {
			return nil, fmt.Errorf("Error: Conjur provider requires the %s environment variable", envvar)
		}
	}

	// Load configuration from the environment
	// TODO get pod namespace / name using Kubernetes API
	// instead of specifying in deployment manifest
	podNamespace := os.Getenv("MY_POD_NAMESPACE")
	podName := os.Getenv("MY_POD_NAME")
	account := os.Getenv("CONJUR_ACCOUNT")
	authnLogin := os.Getenv("CONJUR_AUTHN_LOGIN")
	conjurVersion := os.Getenv("CONJUR_VERSION")
	if len(conjurVersion) == 0 {
		conjurVersion = "5"
	}

	// Load CA cert
	if conjurCACert, err = readSSLCert(); err != nil {
		return nil, err
	}

	// Create new Authenticator
	authenticator, err := authenticator.New(
		authenticator.Config{
			Account:        account,
			ClientCertPath: clientCertPath,
			ConjurVersion:  conjurVersion,
			PodName:        podName,
			PodNamespace:   podNamespace,
			SSLCertificate: conjurCACert,
			TokenFilePath:  tokenFile,
			URL:            authnURL,
			Username:       authnLogin,
		})
	if err != nil {
		return nil, err
	}

	// Send the login request to Conjur to retrieve the signed certificate
	// Configure exponential backoff
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second
	expBackoff.RandomizationFactor = 0.5
	expBackoff.Multiplier = 2
	expBackoff.MaxInterval = 15 * time.Second
	expBackoff.MaxElapsedTime = 2 * time.Minute

	// Try login with exponential backoff on failure
	err = backoff.Retry(
		func() error {
			if err = authenticator.Login(); err != nil {
				return err
			}

			return nil
		},
		expBackoff)

	// If unable to login, return error
	if err != nil {
		return nil, fmt.Errorf("Error: Conjur provider unable to log in to Conjur: %s", err.Error())
	}

	return authenticator, nil
}

// refreshAccessToken uses the Conjur Kubernetes authenticator
// to authenticate with Conjur and retrieve a new time-limited
// access token in a loop
func (p Provider) refreshAccessToken() error {

	var err error

	if p.Authenticator == nil {
		return errors.New("Error: Conjur Kubernetes authenticator must be instantiated before access token may be refreshed")
	}

	// Configure exponential backoff
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second
	expBackoff.RandomizationFactor = 0.5
	expBackoff.Multiplier = 2
	expBackoff.MaxInterval = 15 * time.Second
	expBackoff.MaxElapsedTime = 2 * time.Minute

	// Authenticate in a loop with retries on failure with exponential backoff
	err = backoff.Retry(func() error {
		for {

			// Lock the authenticatorMutex
			p.AuthenticationMutex.Lock()

			log.Printf("Info: Conjur provider is authenticating as %s ...", p.Authenticator.Config.Username)
			resp, err := p.Authenticator.Authenticate()

			if err == nil {
				log.Printf("Info: Conjur provider received a valid authentication response")
				err = p.Authenticator.ParseAuthenticationResponse(resp)
			}

			if err != nil {
				log.Printf("Info: Conjur provider received an error on authenticate: %s", err.Error())

				if autherr, ok := err.(*authenticator.Error); ok {
					if autherr.CertExpired() {
						log.Printf("Info: Conjur certificate expired; Conjur provider is re-logging in.")

						if err = p.Authenticator.Login(); err != nil {
							return err
						}

						// if the cert expired and login worked then continue
						continue
					}
				} else {
					return fmt.Errorf("Error: Conjur provider unable to authenticate: %s", err.Error())
				}
			}

			// Unlock the authenticatorMutex
			p.AuthenticationMutex.Unlock()

			// Reset exponential backoff
			expBackoff.Reset()

			log.Printf("Info: Conjur provider is waiting for %v minutes to re-authenticate.", authenticateCycleDuration)
			time.Sleep(authenticateCycleDuration)
		}
	}, expBackoff)

	if err != nil {
		return fmt.Errorf("Error: Conjur provider unable to authenticate; backoff exhausted: %s", err.Error())
	}

	return nil
}

func readSSLCert() ([]byte, error) {
	SSLCert := os.Getenv("CONJUR_SSL_CERTIFICATE")
	SSLCertPath := os.Getenv("CONJUR_CERT_FILE")
	if SSLCert == "" && SSLCertPath == "" {
		err := errors.New("At least one of CONJUR_SSL_CERTIFICATE and CONJUR_CERT_FILE must be provided")
		return nil, err
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}

	return ioutil.ReadFile(SSLCertPath)
}
