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

	handlerImpl "github.com/conjurinc/secretless/internal/app/secretless/http/handler"
	"github.com/conjurinc/secretless/internal/app/secretless/variable"
	"github.com/conjurinc/secretless/pkg/secretless/config"
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
				serviceType := handler.Type
				if serviceType == "" {
					serviceType = handler.Name
				}
				switch serviceType {
				case "aws":
					return handlerImpl.AWSHandler{handler}
				case "conjur":
					return handlerImpl.ConjurHandler{handler}
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
		http.Error(w, err.Error(), 500)
		return
	}

	if handler.Configuration().Debug {
		log.Printf("Received response %v", resp.Status)
	}

	copyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err := resp.Body.Close(); err != nil {
		log.Printf("Can't close response body %v", err)
	}
}

func (l *Listener) Listen() {
	caCertPool := x509.NewCertPool()
	for _, fname := range l.Config.CACertFiles {
		severCert, err := ioutil.ReadFile(fname)
		if err != nil {
			panic(fmt.Sprintf("Could not load CA certificate file %s : %s", fname, err))
		}
		caCertPool.AppendCertsFromPEM(severCert)
	}

	l.Transport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool}}
	http.Serve(l.Listener, l)
}
