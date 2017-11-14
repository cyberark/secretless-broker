package http

import (
  "io"
  "log"
  "net"
  "net/http"

  "github.com/kgilpin/secretless/pkg/secretless/config"
  "github.com/kgilpin/secretless/internal/app/secretless/variable"
  handlerImpl "github.com/kgilpin/secretless/internal/app/secretless/http/handler"
)

type Handler interface {
  Configuration() *config.Handler

  Authenticate(map[string]string, *http.Request) error
}

type Listener struct {
  Config    config.Listener
  Transport http.Transport
  Handlers  []config.Handler
  Listener  net.Listener
}

// Attribution: https://github.com/elazarl/goproxy/blob/de25c6ed252fdc01e23dae49d6a86742bd790b12/proxy.go#L74
func copyHeaders(dst, src http.Header) {
  for k, _ := range dst {
    dst.Del(k)
  }
  for k, vs := range src {
    for _, v := range vs {
      dst.Add(k, v)
    }
  }
}

func (self *Listener) LookupHandler(r *http.Request) Handler {
  for _, handler := range self.Handlers {
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
func (self *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  //r.Header["X-Forwarded-For"] = w.RemoteAddr()
  if r.Method == "CONNECT" {
    http.Error(w, "CONNECT is not supported.", 405)
    return
  } else {
    var err error

    log.Printf("Got request %v %v %v %v", r.URL.Path, r.Host, r.Method, r.URL.String())

    if !r.URL.IsAbs() {
      http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", 500)
      return
    }

    r.Header.Del("Proxy-Connection")
    r.Header.Del("Proxy-Authenticate")
    r.Header.Del("Proxy-Authorization")

    handler := self.LookupHandler(r)

    if handler != nil {
      if backendVariables, err := variable.Resolve(handler.Configuration().Backend); err != nil {
        http.Error(w, err.Error(), 500)
        return
      } else {
        if handler.Configuration().Debug {
          log.Printf("Backend connection parameters: %s", backendVariables)
        }
        if err = handler.Authenticate(*backendVariables, r); err != nil {
          http.Error(w, err.Error(), 500)
          return
        }
      }
    }

    r.RequestURI = "" // this must be reset when serving a request with the client
    resp, err := self.Transport.RoundTrip(r)
    if err != nil {
      http.Error(w, err.Error(), 500)
      return
    }
    log.Printf("Received response %v", resp.Status)

    copyHeaders(w.Header(), resp.Header)

    w.WriteHeader(resp.StatusCode)

    _, err = io.Copy(w, resp.Body)
    if err := resp.Body.Close(); err != nil {
      log.Printf("Can't close response body %v", err)
    }
  }
}
func (self *Listener) Listen() {
  self.Transport = http.Transport{}
  http.Serve(self.Listener, self)
}
