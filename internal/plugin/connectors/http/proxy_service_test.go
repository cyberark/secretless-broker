package http

import (
	"net"
	gohttp "net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cyberark/secretless-broker/internal/log"
	configv2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	httpconnector "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
)

var _ httpconnector.Connector = &fakeConnector{}

// fakeConnector injects a credential header into the request.
type fakeConnector struct{}

func (f *fakeConnector) Connect(r *gohttp.Request, _ connector.CredentialValuesByID) error {
	r.Header.Set("Authorization", "Basic secret-credential")
	return nil
}

// TestUpstreamURL_BlocksRequestToWrongHost verifies that when upstreamURL is
// configured, requests to a non-matching host are rejected with 403 and
// credentials are never injected or forwarded.
func TestUpstreamURL_BlocksRequestToWrongHost(t *testing.T) {
	// Set up a legitimate upstream server.
	legitimateURL, _ := url.Parse("http://api.legitimate.com:9999")

	// Set up an "attacker" server that should not receive any requests.
	attackerGotRequest := false
	attacker := httptest.NewServer(gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		attackerGotRequest = true
		w.WriteHeader(200)
	}))
	defer attacker.Close()

	// Create a proxy service with upstreamURL pinned to the legitimate server.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	logger := log.New(false)

	subservices := []Subservice{
		{
			ConnectorID: "basic_auth",
			Connector:   &fakeConnector{},
			RetrieveCredentials: func() (map[string][]byte, error) {
				return map[string][]byte{"username": []byte("user"), "password": []byte("pass")}, nil
			},
			AuthenticateURLsMatching: []*regexp.Regexp{regexp.MustCompile(".*")},
			UpstreamURL:              legitimateURL,
		},
	}

	svc, err := NewProxyService(subservices, listener, logger)
	require.NoError(t, err)
	require.NoError(t, svc.Start())
	defer svc.Stop()

	// Client sends a request to the attacker host via the proxy.
	proxyURL, _ := url.Parse("http://" + listener.Addr().String())
	client := &gohttp.Client{
		Transport: &gohttp.Transport{
			Proxy: gohttp.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(attacker.URL + "/steal-creds")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 403, resp.StatusCode)
	assert.False(t, attackerGotRequest, "attacker server must NOT receive the request")
}

// TestUpstreamURL_BlocksRequestWithWrongScheme verifies that a request using
// the correct host but wrong scheme (http vs https) is rejected with 403 and
// credentials are never injected or forwarded. This guards against
// scheme-downgrade attacks where an attacker sends http:// to a service
// configured for https://, aiming to observe credentials in plaintext transit.
func TestUpstreamURL_BlocksRequestWithWrongScheme(t *testing.T) {
	// Set up a server that must NOT receive any request.
	attackerGotRequest := false
	attacker := httptest.NewServer(gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		attackerGotRequest = true
		w.WriteHeader(200)
	}))
	defer attacker.Close()

	// upstreamURL uses https:// for the same host:port as the attacker server.
	// The client will send http://, triggering the scheme mismatch check.
	attackerURL, _ := url.Parse(attacker.URL)
	httpsUpstream, _ := url.Parse("https://" + attackerURL.Host)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	logger := log.New(false)

	subservices := []Subservice{
		{
			ConnectorID: "basic_auth",
			Connector:   &fakeConnector{},
			RetrieveCredentials: func() (map[string][]byte, error) {
				return map[string][]byte{"username": []byte("user"), "password": []byte("pass")}, nil
			},
			AuthenticateURLsMatching: []*regexp.Regexp{regexp.MustCompile(".*")},
			UpstreamURL:              httpsUpstream,
		},
	}

	svc, err := NewProxyService(subservices, listener, logger)
	require.NoError(t, err)
	require.NoError(t, svc.Start())
	defer svc.Stop()

	// Client sends http:// (wrong scheme) to the correct host via the proxy.
	proxyURL, _ := url.Parse("http://" + listener.Addr().String())
	client := &gohttp.Client{
		Transport: &gohttp.Transport{
			Proxy: gohttp.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(attacker.URL + "/steal-creds") // http:// scheme
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 403, resp.StatusCode)
	assert.False(t, attackerGotRequest, "attacker server must NOT receive the request")
}

// TestUpstreamURL_AllowsRequestToCorrectHost verifies that requests targeting
// the configured upstream host are forwarded with credentials injected.
func TestUpstreamURL_AllowsRequestToCorrectHost(t *testing.T) {
	// Set up the legitimate upstream server.
	var receivedAuth string
	var receivedPath string
	legitimate := httptest.NewServer(gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedPath = r.URL.Path
		w.WriteHeader(200)
	}))
	defer legitimate.Close()

	legitimateURL, _ := url.Parse(legitimate.URL)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	logger := log.New(false)

	subservices := []Subservice{
		{
			ConnectorID: "basic_auth",
			Connector:   &fakeConnector{},
			RetrieveCredentials: func() (map[string][]byte, error) {
				return map[string][]byte{"username": []byte("user"), "password": []byte("pass")}, nil
			},
			AuthenticateURLsMatching: []*regexp.Regexp{regexp.MustCompile(".*")},
			UpstreamURL:              legitimateURL,
		},
	}

	svc, err := NewProxyService(subservices, listener, logger)
	require.NoError(t, err)
	require.NoError(t, svc.Start())
	defer svc.Stop()

	// Client sends a request to the correct host via the proxy.
	proxyURL, _ := url.Parse("http://" + listener.Addr().String())
	client := &gohttp.Client{
		Transport: &gohttp.Transport{
			Proxy: gohttp.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(legitimate.URL + "/v1/charges")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Basic secret-credential", receivedAuth)
	assert.Equal(t, "/v1/charges", receivedPath)
}

// TestWithoutUpstreamURL_AttackerReceivesCredentials demonstrates that
// without upstreamURL, credentials are sent to whatever
// host the client specifies.
func TestWithoutUpstreamURL_AttackerReceivesCredentials(t *testing.T) {
	// Set up an "attacker" server.
	var attackerReceivedAuth string
	attacker := httptest.NewServer(gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		attackerReceivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer attacker.Close()

	// Create a proxy service without upstreamURL.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	logger := log.New(false)

	subservices := []Subservice{
		{
			ConnectorID: "basic_auth",
			Connector:   &fakeConnector{},
			RetrieveCredentials: func() (map[string][]byte, error) {
				return map[string][]byte{"username": []byte("user"), "password": []byte("pass")}, nil
			},
			AuthenticateURLsMatching: []*regexp.Regexp{regexp.MustCompile(".*")},
			// UpstreamURL is nil
		},
	}

	svc, err := NewProxyService(subservices, listener, logger)
	require.NoError(t, err)
	require.NoError(t, svc.Start())
	defer svc.Stop()

	// Client sends a request to the attacker host via the proxy.
	proxyURL, _ := url.Parse("http://" + listener.Addr().String())
	client := &gohttp.Client{
		Transport: &gohttp.Transport{
			Proxy: gohttp.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(attacker.URL + "/exfiltrate")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "Basic secret-credential", attackerReceivedAuth)
}

// =============================================================================
// normalizeHost tests
// =============================================================================

// TestNormalizeHost_UnitCases verifies the normalizeHost helper in isolation:
// default ports are appended when absent, explicit ports are left unchanged,
// and the two forms of the same address compare equal.
func TestNormalizeHost_UnitCases(t *testing.T) {
	cases := []struct {
		host     string
		scheme   string
		expected string
	}{
		// HTTPS default port
		{
			"api.example.com",
			"https",
			"api.example.com:443",
		},
		{
			"api.example.com:443",
			"https",
			"api.example.com:443",
		},
		// HTTP default port
		{
			"api.example.com",
			"http",
			"api.example.com:80",
		},
		{
			"api.example.com:80",
			"http",
			"api.example.com:80",
		},
		// Non-default explicit port — left unchanged
		{
			"api.example.com:8443",
			"https",
			"api.example.com:8443",
		},
		{
			"api.example.com:8080",
			"http",
			"api.example.com:8080",
		},
		// Unknown scheme — no port appended
		{
			"api.example.com",
			"ftp",
			"api.example.com",
		},
	}
	for _, tc := range cases {
		got := normalizeHost(tc.host, tc.scheme)
		if got != tc.expected {
			t.Errorf("normalizeHost(%q, %q) = %q; want %q", tc.host, tc.scheme, got, tc.expected)
		}
	}
}

// =============================================================================
// Example config routing tests
// =============================================================================
//
// The tests below verify that the authenticateURLsMatching patterns in the
// example generic connector configs behave correctly: matching the intended
// upstream URL and (for specific-host patterns) rejecting crafted bypass URLs.
//
// No external network connections are made. Config files are loaded from
// examples/generic_connector_configs/ relative to this package directory.
//
// exampleConfigDir is the path to the example generic connector configs.
const exampleConfigDir = "../../../../examples/generic_connector_configs"

// configRoutingCase defines routing assertions for one service in an example config.
type configRoutingCase struct {
	file        string // filename under exampleConfigDir
	service     string // service name within the config
	goodURL     string // URL the pattern MUST match
	bypassURL   string // URL the pattern MUST NOT match; (empty = skip)
	upstreamURL string // expected upstreamURL value (scheme://host) asserted by TestExampleConfigs_UpstreamURLRequired
}

// configRoutingCases is the authoritative table of expected routing behavior.
var configRoutingCases = []configRoutingCase{
	// Patterns missing a $ end-anchor — bypassable via subdomain suffix attack.
	// e.g. ^http[s]*\/api\.stripe\.com* has no $ so "api.stripe.com.attacker.net" matches.
	{
		file:        "stripe_secretless.yml",
		service:     "stripe",
		goodURL:     "https://api.stripe.com/v1/charges",
		bypassURL:   "http://api.stripe.com.attacker.net/exfil",
		upstreamURL: "https://api.stripe.com",
	},
	{
		file:        "stripe_secretless.yml",
		service:     "stripe-account",
		goodURL:     "https://api.stripe.com/v1/charges",
		upstreamURL: "https://api.stripe.com",
	},
	{
		file:        "stripe_secretless.yml",
		service:     "stripe-idempotency",
		goodURL:     "https://api.stripe.com/v1/charges",
		upstreamURL: "https://api.stripe.com",
	},
	{
		file:        "stripe_secretless.yml",
		service:     "stripe-account-idempotency",
		goodURL:     "https://api.stripe.com/v1/charges",
		upstreamURL: "https://api.stripe.com",
	},
	{
		file:        "github_secretless.yml",
		service:     "github",
		goodURL:     "https://api.github.com/repos/user/repo",
		bypassURL:   "http://api.github.attacker.net/exfil",
		upstreamURL: "https://api.github.com",
	},
	{
		file:        "sendgrid_secretless.yml",
		service:     "generic-oauth2",
		goodURL:     "https://api.sendgrid.com/v3/mail/send",
		bypassURL:   "http://api.sendgrid.com.attacker.net/exfil",
		upstreamURL: "https://api.sendgrid.com",
	},
	{
		file:        "twitter_secretless.yml",
		service:     "twitter-oauth2",
		goodURL:     "https://api.twitter.com/2/tweets",
		bypassURL:   "http://api.twitter.com.attacker.net/exfil",
		upstreamURL: "https://api.twitter.com",
	},
	{
		file:        "twitter_secretless.yml",
		service:     "twitter-oauth1",
		goodURL:     "https://api.twitter.com/1.1/statuses/update.json",
		upstreamURL: "https://api.twitter.com",
	},
	// Patterns with properly escaped host and anchored path — NOT host-bypassable via regex.
	// slack: slack\.com\/api* — host portion cannot be substituted; the .com is not a suffix.
	// papertrail: papertrailapp\.com\/api\/v1\/ — specific enough that no attacker host satisfies it.
	// Both still need upstreamURL; without it the broker is still exposed.
	{
		file:        "slack_secretless.yml",
		service:     "slack-form-urlencoded",
		goodURL:     "https://slack.com/api/chat.postMessage",
		upstreamURL: "https://slack.com",
	},
	{
		file:        "slack_secretless.yml",
		service:     "slack-json",
		goodURL:     "https://slack.com/api/chat.postMessage",
		upstreamURL: "https://slack.com",
	},
	{
		file:        "papertrail_secretless.yml",
		service:     "papertrail-token",
		goodURL:     "https://papertrailapp.com/api/v1/events.json",
		upstreamURL: "https://papertrailapp.com",
	},
	{
		file:        "papertrail_secretless.yml",
		service:     "papertrail-basic",
		goodURL:     "https://papertrailapp.com/api/v1/events.json",
		upstreamURL: "https://papertrailapp.com",
	},

	// Broad patterns — no bypassURL.
	{
		file:        "datadog_secretless.yml",
		service:     "datadog-client",
		goodURL:     "https://api.datadoghq.com/api/v1/check_run",
		upstreamURL: "https://api.datadoghq.com",
	},
	{
		file:        "datadog_secretless.yml",
		service:     "datadog-application",
		goodURL:     "https://api.datadoghq.com/api/v1/check_run",
		upstreamURL: "https://api.datadoghq.com",
	},
	{
		file:        "splunk_secretless.yml",
		service:     "splunk",
		goodURL:     "https://splunk.example.com/services/collector",
		upstreamURL: "https://splunk.example.com",
	},
	{
		file:        "new_relic_secretless.yml",
		service:     "new-relic",
		goodURL:     "https://api.newrelic.com/v2/applications.json",
		upstreamURL: "https://api.newrelic.com",
	},
	{
		file:        "google_maps_secretless.yml",
		service:     "google",
		goodURL:     "https://maps.googleapis.com/maps/api/geocode/json",
		upstreamURL: "https://maps.googleapis.com",
	},
	{
		file:        "elasticsearch_secretless.yml",
		service:     "token-service",
		goodURL:     "https://elasticsearch.example.com/_search",
		upstreamURL: "https://elasticsearch.example.com",
	},
	{
		file:        "elasticsearch_secretless.yml",
		service:     "api-key-service",
		goodURL:     "https://elasticsearch.example.com/_search",
		upstreamURL: "https://elasticsearch.example.com",
	},
	{
		file:        "sentry_secretless.yml",
		service:     "sentry-oauth",
		goodURL:     "https://sentry.io/api/0/projects/",
		upstreamURL: "https://sentry.io",
	},
	{
		file:        "loggly_secretless.yml",
		service:     "loggy-token",
		goodURL:     "https://logs-01.loggly.com/inputs/TOKEN/tag/http/",
		upstreamURL: "https://logs-01.loggly.com",
	},
	{
		file:        "loggly_secretless.yml",
		service:     "loggy-basic",
		goodURL:     "https://logs-01.loggly.com/inputs/TOKEN/tag/http/",
		upstreamURL: "https://logs-01.loggly.com",
	},
	{
		file:        "logentries_secretless.yml",
		service:     "logentries",
		goodURL:     "https://data.logentries.com/",
		upstreamURL: "https://data.logentries.com",
	},
	{
		file:        "service_now_secretless.yml",
		service:     "generic-oauth2",
		goodURL:     "https://instance.service-now.com/api/now/table/incident",
		upstreamURL: "https://instance.service-now.com",
	},
	{
		file:        "tableau_secretless.yml",
		service:     "tableau",
		goodURL:     "https://tableau.example.com/api/2.4/auth/signin",
		upstreamURL: "https://tableau.example.com",
	},
	{
		file:        "docker_registry_secretless.yml",
		service:     "docker-registry",
		goodURL:     "https://registry-1.docker.io/v2/",
		upstreamURL: "https://registry-1.docker.io",
	},
	{
		file:        "oauth2_secretless.yml",
		service:     "generic-oauth2",
		goodURL:     "https://api.example.com/resource",
		upstreamURL: "https://api.example.com",
	},
	{
		file:        "instagram_secretless.yml",
		service:     "instagram-oauth2",
		goodURL:     "https://graph.instagram.com/me",
		upstreamURL: "https://graph.instagram.com",
	},
	{
		file:        "dropbox_secretless.yml",
		service:     "dropbox-basic",
		goodURL:     "https://api.dropboxapi.com/2/files/list_folder",
		upstreamURL: "https://api.dropboxapi.com",
	},
	{
		file:        "dropbox_secretless.yml",
		service:     "dropbox-oauth2",
		goodURL:     "https://api.dropboxapi.com/2/files/list_folder",
		upstreamURL: "https://api.dropboxapi.com",
	},
	{
		file:        "dropbox_secretless.yml",
		service:     "dropbox-api-arg",
		goodURL:     "https://api.dropboxapi.com/2/files/download",
		upstreamURL: "https://api.dropboxapi.com",
	},
	{
		file:        "full_contact_secretless.yml",
		service:     "full-contact-oauth2",
		goodURL:     "https://api.fullcontact.com/v3/contacts.search",
		upstreamURL: "https://api.fullcontact.com",
	},
	{
		file:        "full_contact_secretless.yml",
		service:     "full-contact-api-key",
		goodURL:     "https://api.fullcontact.com/v3/contacts.search",
		upstreamURL: "https://api.fullcontact.com",
	},
	{
		file:        "mailchimp_secretless.yml",
		service:     "mailchimp-basic-auth",
		goodURL:     "https://us1.api.mailchimp.com/3.0/lists",
		upstreamURL: "https://us1.api.mailchimp.com",
	},
	{
		file:        "mailchimp_secretless.yml",
		service:     "mailchimp-oauth2",
		goodURL:     "https://us1.api.mailchimp.com/3.0/lists",
		upstreamURL: "https://us1.api.mailchimp.com",
	},
	{
		file:        "twilio_secretless.yml",
		service:     "twilio",
		goodURL:     "https://api.twilio.com/2010-04-01/Accounts/",
		upstreamURL: "https://api.twilio.com",
	},
	{
		file:        "jfrog_artifactory_secretless.yml",
		service:     "jfrog-basic-authentication",
		goodURL:     "https://mycompany.jfrog.io/artifactory/api/",
		upstreamURL: "https://mycompany.jfrog.io",
	},
	{
		file:        "jfrog_artifactory_secretless.yml",
		service:     "jfrog-apikey",
		goodURL:     "https://mycompany.jfrog.io/artifactory/api/",
		upstreamURL: "https://mycompany.jfrog.io",
	},
	{
		file:        "jfrog_artifactory_secretless.yml",
		service:     "jfrog-oauth",
		goodURL:     "https://mycompany.jfrog.io/artifactory/api/",
		upstreamURL: "https://mycompany.jfrog.io",
	},
	{
		file:        "oauth1_secretless.yml",
		service:     "generic-oauth1",
		goodURL:     "https://api.example.com/1.0/resource",
		upstreamURL: "https://api.example.com",
	},
}

// loadExampleHTTPConfig parses an example config file and returns the
// compiled HTTPConfig for the named service. Fails the test immediately on
// any parse error.
//
// Note: configv2.NewHTTPConfig currently panics on invalid regexp syntax
// instead of returning an error (a known bug). None of the current example
// configs trigger this, but fixing that panic-to-error is a prerequisite for
// safe future use of this helper.
func loadExampleHTTPConfig(t *testing.T, filename, serviceName string) *configv2.HTTPConfig {
	t.Helper()
	data, err := os.ReadFile(exampleConfigDir + "/" + filename)
	require.NoError(t, err, "reading %s", filename)

	cfg, err := configv2.NewConfig(data)
	require.NoError(t, err, "parsing %s", filename)

	for _, svc := range cfg.Services {
		if svc.Name == serviceName {
			httpCfg, err := configv2.NewHTTPConfig([]byte(svc.ConnectorConfig))
			require.NoError(t, err, "parsing HTTPConfig for service %q in %s", serviceName, filename)
			require.NotEmpty(t, httpCfg.AuthenticateURLsMatching,
				"service %q in %s must have at least one authenticateURLsMatching pattern", serviceName, filename)
			return httpCfg
		}
	}

	t.Fatalf("service %q not found in %s", serviceName, filename)
	return nil
}

func matchesAnyPattern(patterns []*regexp.Regexp, rawURL string) bool {
	for _, p := range patterns {
		if p.MatchString(rawURL) {
			return true
		}
	}
	return false
}

// TestExampleConfigs_Load verifies that every example config file parses
// without error via v2.NewConfig: YAML structure, regexp compilation, and
// upstreamURL URL validation. No network connections are made.
func TestExampleConfigs_Load(t *testing.T) {
	seen := map[string]bool{}
	for _, tc := range configRoutingCases {
		if seen[tc.file] {
			continue
		}
		seen[tc.file] = true
		tc := tc
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(exampleConfigDir + "/" + tc.file)
			require.NoError(t, err)
			_, err = configv2.NewConfig(data)
			assert.NoError(t, err)
		})
	}
}

// TestExampleConfigs_RegexRouting asserts positive and negative routing
// behavior for every service in configRoutingCases. No network connections.
//
// Positive assertion (all configs): the pattern must match goodURL. A
// failure here means the config rejects its own legitimate traffic.
//
// Negative assertion (specific-host configs only): the pattern must NOT match
// bypassURL.
func TestExampleConfigs_RegexRouting(t *testing.T) {
	for _, tc := range configRoutingCases {
		tc := tc
		t.Run(tc.file+"/"+tc.service, func(t *testing.T) {
			httpCfg := loadExampleHTTPConfig(t, tc.file, tc.service)

			assert.True(t,
				matchesAnyPattern(httpCfg.AuthenticateURLsMatching, tc.goodURL),
				"patterns %v should match intended URL %q",
				httpCfg.AuthenticateURLsMatching, tc.goodURL,
			)

			if tc.bypassURL != "" {
				assert.False(t,
					matchesAnyPattern(httpCfg.AuthenticateURLsMatching, tc.bypassURL),
					"SECURITY: patterns %v must NOT match bypass URL %q — "+
						"fix: escape dots (\\.), use https:// prefix, anchor end with ($|/)",
					httpCfg.AuthenticateURLsMatching, tc.bypassURL,
				)
			}
		})
	}
}

// TestExampleConfigs_UpstreamURLRequired asserts that every example config
// service has upstreamURL set.
func TestExampleConfigs_UpstreamURLRequired(t *testing.T) {
	for _, tc := range configRoutingCases {
		tc := tc
		t.Run(tc.file+"/"+tc.service, func(t *testing.T) {
			httpCfg := loadExampleHTTPConfig(t, tc.file, tc.service)
			assert.NotNil(t, httpCfg.UpstreamURL,
				"Service %q in %s has no upstreamURL. "+
					"Credentials will be forwarded to any host the client specifies. "+
					"Add: upstreamURL: \"https://<canonical-upstream-host>\"",
				tc.service, tc.file,
			)
		})
	}
}
