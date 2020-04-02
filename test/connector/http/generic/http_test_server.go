package generic

import (
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// testServer creates, for testing purposes, an http server with basic-auth on a random
// port.
func testServer(
	serverUsername string,
	serverPassword string,
) (*httptest.Server, error) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		notAuthenticated := !ok ||
			subtle.ConstantTimeCompare([]byte(user), []byte(serverUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(serverPassword)) != 1

		if notAuthenticated {
			http.Error(w, serverResponseUnauthorized, http.StatusUnauthorized)
			return
		}

		_, _ = fmt.Fprintln(w, serverResponseOK)
	})

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	ts := &httptest.Server{
		Listener: listener,
		Config: &http.Server{
			Handler: handler,
		},
	}

	return ts, nil
}

// targetEndpoint gets the host and port for a test server.
func targetEndpoint(srv *httptest.Server) string {
	u, _ := url.Parse(srv.URL)
	return serverHostname + ":" + u.Port()
}

// httpServer creates a test server using the basic auth credentials passed in as
// arguments.
func httpServer(
	serverUsername string,
	serverPassword string,
) (*httptest.Server, error) {
	s, err := testServer(serverUsername, serverPassword)
	if err != nil {
		return nil, err
	}

	s.Start()
	return s, nil
}

// httpsServer does exactly what httpServer but allows the server to run with TLS enabled.
// The TLS key pair are provided as arguments.
func httpsServer(
	serverUsername string,
	serverPassword string,
	tlsCert string,
	tlsKey string,
) (*httptest.Server, error) {
	s, err := testServer(serverUsername, serverPassword)
	if err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	s.TLS = config

	s.StartTLS()
	return s, nil
}

// proxyGet is a convenience method that makes an HTTP GET request using a proxy
func proxyGet(endpoint, proxy string) (*http.Response, error) {
	req, err := http.NewRequest(
		"GET",
		endpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{Proxy: func(req *http.Request) (proxyURL *url.URL, err error) {
		return url.Parse(proxy)
	}}
	client := &http.Client{Transport: transport}
	return client.Do(req)
}
