package ssl

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testCertificates is used to store all the test certificates
type testCertificates struct {
	serverCert []byte
	serverKey  []byte
	rootCert   []byte
	clientCert []byte
	clientKey  []byte
}

// loadTestCerts loads test certificates from the `./testdata` directory
func loadTestCerts() (*testCertificates, error) {
	serverCert, err := ioutil.ReadFile("./testdata/server.pem")
	if err != nil {
		return nil, err
	}
	serverKey, err := ioutil.ReadFile("./testdata/server-key.pem")
	if err != nil {
		return nil, err
	}
	rootCert, err := ioutil.ReadFile("./testdata/ca.pem")
	if err != nil {
		return nil, err
	}
	clientCert, err := ioutil.ReadFile("./testdata/client.pem")
	if err != nil {
		return nil, err
	}
	clientKey, err := ioutil.ReadFile("./testdata/client-key.pem")
	if err != nil {
		return nil, err
	}

	return &testCertificates{
		serverCert: serverCert,
		serverKey:  serverKey,
		rootCert:   rootCert,
		clientCert: clientCert,
		clientKey:  clientKey,
	}, nil
}

// httpsTestServer is a HTTP test server with TLS. It's a light wrapper around the
// server you get from the httptest package. It's very convenient to use.
func httpsTestServer(
	serverCert []byte,
	serverKey []byte,
) (*httptest.Server, error) {
	cert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintln(w, "Hello, client")
		}))

	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	ts.StartTLS()

	return ts, nil
}

func TestHandleSSLUpgrade(t *testing.T) {
	// Load test certificates
	testCerts, err := loadTestCerts()
	if !assert.NoError(t, err) {
		return
	}

	// Run the HTTP test server with TLS
	ts, err := httpsTestServer(
		testCerts.serverCert,
		testCerts.serverKey,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer ts.Close()

	// Create sslmode with verify-ca for the test because it exercise most of the ssl
	// package.
	sslmode, err := NewDbSSLMode(
		options{
			"host":        "localhost",
			"sslmode":     "verify-ca",
			"sslrootcert": string(testCerts.rootCert),
			"sslcert":     string(testCerts.clientCert),
			"sslkey":      string(testCerts.clientKey),
		}, false)
	if !assert.NoError(t, err) {
		return
	}

	// Dial to the test server
	conn, err := net.Dial(
		ts.Listener.Addr().Network(),
		ts.Listener.Addr().String(),
	)
	if !assert.NoError(t, err) {
		return
	}

	// Upgrade connection using sslmode
	upgradedConn, err := HandleSSLUpgrade(conn, sslmode)
	if !assert.NoError(t, err) {
		return
	}
	// Ensure that the upgraded connection is a TLS connection
	assert.IsType(t, upgradedConn, &tls.Conn{})
}

func TestNewDbSSLMode_valid(t *testing.T) {
	// validSSLModeTestCase represents tests cases for NewDbSSLMode when the sslmode option
	// is a valid value such as 'require'. The tests make assertions on the resulting
	// DbSSLMode from NewDbSSLMode and anticipate no error.
	type validSSLModeTestCase struct {
		description string
		options     options
		assertion   func(t *testing.T, sslmode DbSSLMode)
	}

	var validSSLModeTestCases = []validSSLModeTestCase{
		{
			description: "sslmode=disable",
			options:     options{"sslmode": "disable"},
			assertion: func(t *testing.T, sslmode DbSSLMode) {
				assert.False(t, sslmode.UseTLS)
			},
		},
		{
			description: "sslmode=required",
			options:     options{"sslmode": "require"},
			assertion: func(t *testing.T, sslmode DbSSLMode) {
				assert.True(t, sslmode.UseTLS)
				assert.False(t, sslmode.VerifyCaOnly)
			},
		},
		{
			description: "sslmode=verify-ca",
			options:     options{"sslmode": "verify-ca"},
			assertion: func(t *testing.T, sslmode DbSSLMode) {
				assert.True(t, sslmode.UseTLS)
				assert.True(t, sslmode.VerifyCaOnly)
			},
		},
		{
			description: "sslmode=verify-full",
			options: options{
				"sslmode": "verify-full",
				"host":    "some-host",
			},
			assertion: func(t *testing.T, sslmode DbSSLMode) {
				assert.True(t, sslmode.UseTLS)
				assert.Equal(t, sslmode.ServerName, "some-host")
			},
		},
		{
			description: "sslmode=verify-full sslhost takes precedence",
			options: options{
				"sslmode": "verify-full",
				"host":    "some-host",
				"sslhost": "overridden-host",
			},
			assertion: func(t *testing.T, sslmode DbSSLMode) {
				assert.True(t, sslmode.UseTLS)
				assert.Equal(t, sslmode.ServerName, "overridden-host")
			},
		},
	}

	for _, testCase := range validSSLModeTestCases {
		t.Run(testCase.description, func(t *testing.T) {
			sslmode, err := NewDbSSLMode(testCase.options, false)
			if !assert.NoError(t, err) {
				return
			}

			testCase.assertion(t, sslmode)
		})
	}
}

func TestNewDbSSLMode(t *testing.T) {
	t.Run("Options are passed as is", func(t *testing.T) {
		opts := options{
			"a": "b",
			"x": "y",
		}

		sslmode, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, sslmode.Options, opts)
	})

	t.Run("Invalid sslmode option", func(t *testing.T) {
		opts := options{
			"sslmode": "invalid",
		}

		_, err := NewDbSSLMode(
			opts,
			false,
		)
		if !assert.Error(t, err) {
			return
		}
	})
}
