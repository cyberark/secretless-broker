package generic

import (
	"io/ioutil"
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

			body, err := ioutil.ReadAll(res.Body)
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

			body, err := ioutil.ReadAll(res.Body)
			if !assert.NoError(t, err) {
				return
			}

			assert.Contains(t, string(body), testCase.expected)
		})
	}
}
