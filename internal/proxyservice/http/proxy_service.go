package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	gohttp "net/http"

	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	http "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
)

// Subservice handles specific traffic within an HTTP proxy service, using
// traffic filtering rules and a devoted Connector.
type Subservice struct {
	Connector           http.Connector
	RetrieveCredentials internal.CredentialsRetriever
}

// NewProxyService create a new HTTP proxy service.
func NewProxyService(
	subservices []Subservice,
	sharedListener net.Listener,
	logger log.Logger,
) (internal.Service, error) {
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

	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		logger.Warnf("Error '%s' loading system cert pool; will use an empty cert pool", err)
		caCertPool = x509.NewCertPool()
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

func (proxy *proxyService) LookupSubservice(
	r *gohttp.Request,
) *Subservice {
	return nil
}

// ServeHTTP exists to implement the go_http.Handler interface
func (proxy *proxyService) ServeHTTP(w gohttp.ResponseWriter, r *gohttp.Request) {
	logger := proxy.logger

	if r.Method == "CONNECT" {
		gohttp.Error(w, "CONNECT is not supported.", 405)
		return
	}

	var err error

	logger.Infoln("Got request %v %v %v %v", r.URL.Path, r.Host, r.Method, r.URL.String())

	if !r.URL.IsAbs() {
		gohttp.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", 500)
		return
	}

	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")

	subservice := proxy.LookupSubservice(r)

	if subservice != nil {
		creds, err := subservice.RetrieveCredentials()
		defer internal.zeroizeCredentials(creds)
		if err != nil {
			gohttp.Error(w, err.Error(), 500)
			return
		}

		err = subservice.Connector(r, creds)
		if err != nil {
			gohttp.Error(w, err.Error(), 500)
			return
		}

	} else {
		logger.Warnf("No subservice for request: %s %s", r.Method, r.URL)
	}

	r.RequestURI = "" // this must be reset when serving a request with the client

	resp, err := proxy.transport.RoundTrip(r)
	if err != nil {
		logger.Debugf("Error: %v\n", err)
		gohttp.Error(w, err.Error(), 503)
		return
	}

	logger.Debugf("Received response status: %s\n", resp.Status)

	copyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err := resp.Body.Close(); err != nil {
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
