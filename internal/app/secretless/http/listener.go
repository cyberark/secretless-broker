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
	"strconv"

	handlerImpl "github.com/conjurinc/secretless/internal/app/secretless/http/handler"
	"github.com/conjurinc/secretless/internal/app/secretless/variable"
	"github.com/conjurinc/secretless/pkg/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	validation "github.com/go-ozzo/ozzo-validation"
)

type Handler interface {
	Configuration() *config.Handler

	Authenticate(map[string][]byte, *http.Request) error
}

type Listener struct {
	Config    config.Listener
	Transport *http.Transport
	Handlers  []config.Handler
	Listener  net.Listener
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
		}
	}
	return errors.Filter()
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Handlers, validation.Required),
		validation.Field(&l.Handlers, handlerHasCredentials{}),
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

func (l *Listener) LookupHandler(r *http.Request) Handler {
	for _, handler := range l.Handlers {
		for _, pattern := range handler.Patterns {
			log.Printf("Matching handler pattern %s to request %s", pattern.String(), r.URL)
			if pattern.MatchString(r.URL.String()) {
				if handler.Debug {
					log.Printf("Using handler '%s' for request %s", handler.Name, r.URL.String())
				}
				// Construct the return object
				switch handler.Type {
				case "aws":
					return handlerImpl.AWSHandler{handler}
				case "conjur":
					return handlerImpl.ConjurHandler{handler}
				default:
					log.Panicf("Service type '%s' of handler '%s' is not recognized", handler.Type, handler.Name)
				}
			}
		}
	}
	return nil
}

// Standard net/http function. Shouldn't be used directly, http.Serve will use it.
func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		if backendVariables, err = variable.Resolve(handler.Configuration().Credentials); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if handler.Configuration().Debug {
			log.Printf("%s backend connection parameters: %s", handler.Configuration().Name, backendVariables)
		}
		if err = handler.Authenticate(backendVariables, r); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	r.RequestURI = "" // this must be reset when serving a request with the client

	if handler.Configuration().Debug {
		log.Printf("Sending request %v", r)
	}

	resp, err := l.Transport.RoundTrip(r)

	if err != nil {
		if handler.Configuration().Debug {
			log.Printf("Error: %v", err)
		}
		http.Error(w, err.Error(), 500)
		return
	}

	// Note: resp is likely nil if err is non-nil, so don't access it until you get here.

	if handler.Configuration().Debug {
		log.Printf("Received response status: %d", resp.Status)
	}

	copyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err := resp.Body.Close(); err != nil {
		log.Printf("Can't close response body %v", err)
	}
}

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
	http.Serve(l.Listener, l)
}

// GetConfig implements secretless.Listener
func (l *Listener) GetConfig() config.Listener {
	return l.Config
}

// GetListener implements secretless.Listener
func (l *Listener) GetListener() net.Listener {
	return l.Listener
}

// GetHandlers implements secretless.Listener
func (l *Listener) GetHandlers() []secretless.Handler {
	return nil
}

// GetConnections implements secretless.Listener
func (l *Listener) GetConnections() []net.Conn {
	return nil
}
