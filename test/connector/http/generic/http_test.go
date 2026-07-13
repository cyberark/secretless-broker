package generic

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreds(t *testing.T) {
	testCases := []struct {
		description    string
		serverUsername string
		serverPassword string
		expected       string
	}{
		{
			"proxy credentials match server credentials",
			fromProxyUsername,
			fromProxyPassword,
			serverResponseOK,
		},
		{
			"proxy credentials don't match server credentials",
			"not-proxy-user",
			"not-proxy-password",
			serverResponseUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv, err := httpServer(testCase.serverUsername, testCase.serverPassword)
			if !assert.NoError(t, err) {
				return
			}

			defer srv.Close()

			res, err := proxyGet("http://"+targetEndpoint(srv), proxyHTTP)
			if !assert.NoError(t, err) {
				return
			}

			body, err := io.ReadAll(res.Body)
			if !assert.NoError(t, err) {
				return
			}

			assert.Contains(t, string(body), testCase.expected)
		})
	}
}

func TestForceSSL(t *testing.T) {
	testCases := []struct {
		description string
		tlsCert     string
		tlsKey      string
		expected    string
	}{
		{
			"certificate included in proxy bundle",
			serverCertIncluded,
			serverKeyIncluded,
			serverResponseOK,
		},
		{
			"certificate not included proxy bundle",
			serverCertExcluded,
			serverKeyExcluded,
			"x509: certificate signed by unknown authority",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv, err := httpsServer(
				fromProxyUsername,
				fromProxyPassword,
				testCase.tlsCert,
				testCase.tlsKey,
			)
			if !assert.NoError(t, err) {
				return
			}

			defer srv.Close()

			res, err := proxyGet("http://"+targetEndpoint(srv), proxyHTTPS)
			if !assert.NoError(t, err) {
				return
			}

			body, err := io.ReadAll(res.Body)
			if !assert.NoError(t, err) {
				return
			}

			assert.Contains(t, string(body), testCase.expected)
		})
	}
}

// TestForceSSLDefaultRejectsPlaintext verifies that a connector which does not
// set forceSSL uses the secure default (forceSSL: true) and therefore rejects
// requests to a plaintext http upstream instead of forwarding the injected
// credentials in cleartext. The connector upgrades the scheme to https, so the
// plaintext upstream fails the TLS handshake and the proxy returns an error
// rather than a successful, credentialed response.
func TestForceSSLDefaultRejectsPlaintext(t *testing.T) {
	srv, err := httpServer(fromProxyUsername, fromProxyPassword)
	if !assert.NoError(t, err) {
		return
	}
	defer srv.Close()

	res, err := proxyGet("http://"+targetEndpoint(srv), proxyDefault)
	if err != nil {
		return
	}

	assert.NotEqual(t, 200, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotContains(t, string(body), serverResponseOK)
}
