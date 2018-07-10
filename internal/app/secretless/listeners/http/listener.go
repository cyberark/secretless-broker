package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"

	"github.com/conjurinc/secretless/internal/app/secretless/variable"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Listener listens for and handles new connections
type Listener struct {
	Config         config.Listener
	EventNotifier  plugin_v1.EventNotifier
	HandlerConfigs []config.Handler
	NetListener    net.Listener
	RunHandlerFunc func(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler
	Transport      *http.Transport
}

// HandlerHasCredentials validates that a handler has all necessary credentials.
type handlerHasCredentials struct {
}

// Validate checks that a handler has all necessary credentials.
func (hhc handlerHasCredentials) Validate(value interface{}) error {
	hs := value.([]config.Handler)
	errors := validation.Errors{}
	for i, h := range hs {
		if h.Type == "aws" {
			if !h.HasCredential("accessKeyId") {
				errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'accessKeyId'")
			}
			if !h.HasCredential("secretAccessKey") {
				errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'secretAccessKey'")
			}
		} else if h.Type == "conjur" {
			if !h.HasCredential("accessToken") {
				errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'accessToken'")
			}
		} else if h.Type == "basic_auth" {
			for _, credential := range [...]string{"username", "password"} {
				if !h.HasCredential(credential) {
					errors[strconv.Itoa(i)] = fmt.Errorf("must have credential '" + credential + "'")
				}
			}
		} else {
			errors[strconv.Itoa(i)] = fmt.Errorf("Handler type is not supported")
		}
	}

	return errors.Filter()
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.HandlerConfigs, validation.Required),
		validation.Field(&l.HandlerConfigs, handlerHasCredentials{}),
	)
}

// Attribution: https://github.com/elazarl/goproxy/blob/de25c6ed252fdc01e23dae49d6a86742bd790b12/proxy.go#L74
func copyHeaders(dst, src http.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

// LookupHandler returns the handler that matches the request URL
func (l *Listener) LookupHandler(r *http.Request) plugin_v1.Handler {
	for _, handlerConfig := range l.HandlerConfigs {
		for _, pattern := range handlerConfig.Patterns {
			log.Printf("Matching handler pattern %s to request %s", pattern.String(), r.URL)
			if pattern.MatchString(r.URL.String()) {
				if handlerConfig.Debug {
					log.Printf("Using handler '%s' for request %s", handlerConfig.Name, r.URL.String())
				}

				handlerOptions := plugin_v1.HandlerOptions{
					HandlerConfig: handlerConfig,
				}

				return l.RunHandlerFunc("http/"+handlerConfig.Type, handlerOptions)
			}
		}
	}
	return nil
}

// Standard net/http function. Shouldn't be used directly, http.Serve will use it.
func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	listenerDebug := l.Config.Debug

	//r.Header["X-Forwarded-For"] = w.RemoteAddr()
	if r.Method == "CONNECT" {
		http.Error(w, "CONNECT is not supported.", 405)
		return
	}

	var err error

	log.Printf("Got request %v %v %v %v", r.URL.Path, r.Host, r.Method, r.URL.String())

	if !r.URL.IsAbs() {
		http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", 500)
		return
	}

	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")

	handler := l.LookupHandler(r)

	if handler != nil {
		var backendVariables map[string][]byte
		if backendVariables, err = variable.Resolve(handler.GetConfig().Credentials, l.EventNotifier); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if listenerDebug || handler.GetConfig().Debug {
			keys := reflect.ValueOf(backendVariables).MapKeys()
			log.Printf("%s backend connection parameters: %s", handler.GetConfig().Name, keys)
		}
		if err = handler.Authenticate(backendVariables, r); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		log.Printf("WARN: Have no handler for request: %s %s", r.Method, r.URL)
	}

	handlerDebug := handler != nil && handler.GetConfig().Debug

	r.RequestURI = "" // this must be reset when serving a request with the client

	if listenerDebug || handlerDebug {
		log.Printf("Sending request %v", r)
	}

	resp, err := l.Transport.RoundTrip(r)
	if err != nil {
		if listenerDebug || handlerDebug {
			log.Printf("Error: %v", err)
		}
		http.Error(w, err.Error(), 503)
		return
	}

	// Note: resp is likely nil if err is non-nil, so don't access it until you get here.

	if listenerDebug || handlerDebug {
		log.Printf("Received response status: %d", resp.Status)
	}

	copyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err := resp.Body.Close(); err != nil {
		log.Printf("Can't close response body %v", err)
	}
}

// Listen serves HTTP requests
func (l *Listener) Listen() {
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("Error '%s' loading system cert pool; will use an empty cert pool", err)
		caCertPool = x509.NewCertPool()
	}
	for _, fname := range l.Config.CACertFiles {
		severCert, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Panicf("Could not load CA certificate file %s : %s", fname, err)
		}
		caCertPool.AppendCertsFromPEM(severCert)
	}

	l.Transport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool}}
	http.Serve(l.NetListener, l)
}

// GetConfig implements plugin_v1.Listener
func (l *Listener) GetConfig() config.Listener {
	return l.Config
}

// GetListener implements plugin_v1.Listener
func (l *Listener) GetListener() net.Listener {
	return l.NetListener
}

// GetHandlers implements plugin_v1.Listener
func (l *Listener) GetHandlers() []plugin_v1.Handler {
	return nil
}

// GetConnections implements plugin_v1.Listener
func (l *Listener) GetConnections() []net.Conn {
	return nil
}

// GetNotifier implements plugin_v1.Listener
func (l *Listener) GetNotifier() plugin_v1.EventNotifier {
	return l.EventNotifier
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "http"
}

// Shutdown implements plugin_v1.Listener
func (l *Listener) Shutdown() error {
	// TODO: Clean up all handlers
	return l.NetListener.Close()
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	return &Listener{
		Config:         options.ListenerConfig,
		EventNotifier:  options.EventNotifier,
		HandlerConfigs: options.HandlerConfigs,
		NetListener:    options.NetListener,
		RunHandlerFunc: options.RunHandlerFunc,
	}
}
