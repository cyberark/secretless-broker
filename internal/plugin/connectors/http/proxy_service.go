package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	gohttp "net/http"
	"os"
	"regexp"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
)

// Subservice handles specific traffic within an HTTP proxy service, using
// traffic filtering rules and a devoted Connector.
type Subservice struct {

	// NOTE: This existence of both "ConnectorID" and "Authenticate" here
	// indicates a deeper problem: The concept of "connector" probably should
	// have included both ID and "connector function" together, as a single
	// entity.  That feels like the right abstraction, though the costs of not
	// having it are minimal so far. This is something we should keep an eye on
	// and refactor if it comes up again.

	Connector                http.Connector
	ConnectorID              string
	RetrieveCredentials      internal.CredentialsRetriever
	AuthenticateURLsMatching []*regexp.Regexp
}

// Matches returns true if any of the patterns in the Subservice's
// AuthenticateURLsMatching match the given url.
func (sub *Subservice) Matches(url string) bool {
	for _, pattern := range sub.AuthenticateURLsMatching {
		if pattern.MatchString(url) {
			return true
		}
	}
	return false
}

// NewProxyService create a new HTTP proxy service.
func NewProxyService(
	subservices []Subservice,
	sharedListener net.Listener,
	logger log.Logger,
) (internal.Service, error) {

	// Parameter validation

	errors := validation.Errors{}
	if len(subservices) == 0 {
		errors["subservices"] = fmt.Errorf("subservices cannot be empty")
	}
	if sharedListener == nil {
		errors["sharedListener"] = fmt.Errorf("sharedListener cannot be nil")
	}
	if logger == nil {
		errors["logger"] = fmt.Errorf("logger cannot be nil")
	}
	if err := errors.Filter(); err != nil {
		return nil, err
	}
	// prevents IDE "possible nil dereference" warnings below -- better way?
	logger = logger.(log.Logger)

	// Create the http.Transport

	// TODO: Explanation of why we have to do this.  Add ability for user
	//   to override the default pool.
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		msg := "Error '%s' loading system cert pool; using an empty cert pool"
		logger.Warnf(msg, err)
		caCertPool = x509.NewCertPool()
	}

	if caBundle, ok := os.LookupEnv("SECRETLESS_HTTP_CA_BUNDLE"); ok {
		// Read in the cert file
		certs, err := ioutil.ReadFile(caBundle)
		if err != nil {
			return nil, fmt.Errorf("failed to append SECRETLESS_HTTP_CA_BUNDLE to RootCAs: %v", err)
		}

		// Append our cert to the system pool
		if ok := caCertPool.AppendCertsFromPEM(certs); !ok {
			logger.Warnf("No certs appended, using system certs only")
		}
	}

	transport := &gohttp.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	return &proxyService{
		transport:   transport,
		subservices: subservices,
		listener:    sharedListener,
		logger:      logger,
		done:        false,
	}, nil
}

type proxyService struct {
	transport   *gohttp.Transport
	done        bool
	listener    net.Listener
	logger      log.Logger
	subservices []Subservice
}

func (proxy *proxyService) matchingSubservices(r *gohttp.Request) []Subservice {
	var matchingSubs []Subservice
	for _, sub := range proxy.subservices {
		if sub.Matches(r.URL.String()) {
			matchingSubs = append(matchingSubs, sub)
		}
	}
	return matchingSubs
}

// selectSubservice finds a subservice matching the request, and issues warnings
// when there are no matches or more than one match.
func (proxy *proxyService) selectSubservice(r *gohttp.Request) *Subservice {
	matchingSubs := proxy.matchingSubservices(r)

	// No match: Warn!
	if len(matchingSubs) == 0 {
		msg := "No subservices matched request '%s'"
		proxy.logger.Warnf(msg, r.URL.Host)
		return nil
	}

	// Multiple matches: Warn!
	if len(matchingSubs) > 1 {
		msg := "Multiple subservices matched request '%s': %v\n"
		proxy.logger.Warnf(msg, r.URL.Host, matchingSubs)
	}

	// Select first (or only) match
	subservice := matchingSubs[0]
	msg := "Using connector '%s' for request %s"
	proxy.logger.Debugf(msg, subservice.ConnectorID, r.URL.Host)
	return &subservice
}

// ServeHTTP exists to implement the go_http.Handler interface
func (proxy *proxyService) ServeHTTP(w gohttp.ResponseWriter, r *gohttp.Request) {
	logger := proxy.logger

	// Log request
	logMsg := "Got request %v %v %v %v"
	logger.Debugf(logMsg, r.URL.Path, r.Host, r.Method, r.URL.Hostname())

	// Validate request

	if !proxy.validateProxyServerRules(w, r) {
		return
	}

	// Remove headers intended only for proxy server.

	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")

	// Select a subservice

	subservice := proxy.selectSubservice(r)

	// No match: Send request with no authentication.
	if subservice == nil {
		proxy.handleRequest(w, r)
		return
	}

	// Get current credential values

	creds, err := subservice.RetrieveCredentials()
	defer internal.ZeroizeCredentials(creds)

	if err != nil {
		gohttp.Error(w, err.Error(), 500)
		return
	}

	// Authenticate request

	err = subservice.Connector.Connect(r, creds)
	if err != nil {
		gohttp.Error(w, err.Error(), 500)
		return
	}

	// Send request to target service

	proxy.handleRequest(w, r)
}

// validateProxyServerRules ensures that the request being made is a valid
// request for a proxy server.
func (proxy *proxyService) validateProxyServerRules(
	w gohttp.ResponseWriter, r *gohttp.Request,
) bool {

	if r.Method == "CONNECT" {
		gohttp.Error(w, "CONNECT is not supported.", 405)
		return false
	}

	if !r.URL.IsAbs() {
		errMsg := "This is a proxy server. Non-proxy requests aren't allowed."
		gohttp.Error(w, errMsg, 500)
		return false
	}

	return true
}

// handleRequest sends the request to the target service and writes the response
// back to the client.
func (proxy *proxyService) handleRequest(
	w gohttp.ResponseWriter, r *gohttp.Request,
) {
	logger := proxy.logger

	// Per the stdlib docs, "It is an error to set this field in an HTTP client
	// request". Therefore, we ensure it is empty in case the client set it.
	r.RequestURI = ""

	// Send request to target service

	resp, err := proxy.transport.RoundTrip(r)
	if err != nil {
		logger.Debugf("Error: %v\n", err)
		gohttp.Error(w, err.Error(), 503)
		return
	}

	// Send response to client (everything below)

	logger.Debugf("Received response status: %s\n", resp.Status)

	copyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Errorf("Can't write response to body: %s\n", err)
	}

	err = resp.Body.Close()
	if err != nil {
		logger.Debugf("Can't close response body %v\n", err)
	}
}

// Start initiates the net.Listener to listen for incoming connections
func (proxy *proxyService) Start() error {
	logger := proxy.logger

	logger.Infof("Starting service")

	if proxy.done {
		return fmt.Errorf("cannot call Start on stopped ProxyService")
	}

	// We need a Go routine here because http.Serve() is blocking, but this
	// Start() method shouldn't be.
	go func() {
		err := gohttp.Serve(proxy.listener, proxy)
		if err != nil && !proxy.done {
			logger.Errorf("proxy service failed on server: %s", err)
			return
		}
	}()

	return nil
}

// Stop terminates proxyService by closing the listening net.Listener
func (proxy *proxyService) Stop() error {
	proxy.logger.Infof("Stopping service")
	proxy.done = true
	return proxy.listener.Close()
}

// Attribution: https://github.com/elazarl/goproxy/blob/de25c6ed252fdc01e23dae49d6a86742bd790b12/proxy.go#L74
func copyHeaders(dst, src gohttp.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}
