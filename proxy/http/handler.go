package http

import (
  "io"
  "log"
  "net"
  "net/http"

  "github.com/kgilpin/secretless/config"
  "github.com/kgilpin/secretless/variable"
)

type Authenticator interface {
  Authenticate(map[string]string, *http.Request) error
}

type HTTPHandler struct {
  Config        config.ListenerConfig
  Transport     *http.Transport
  Authenticator Authenticator
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

// Standard net/http function. Shouldn't be used directly, http.Serve will use it.
func (self *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	  if backendVariables, err := variable.Resolve(self.Config.Backend); err != nil {
      http.Error(w, err.Error(), 500)
	    return
	  } else {
	    if err = self.Authenticator.Authenticate(*backendVariables, r); err != nil {
	      http.Error(w, err.Error(), 500)
	      return
	    }
	  }

	  if self.Config.Debug {
	    log.Printf("Header after authentication processing: %v", r.Header)
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

func (self *HTTPHandler) Run(l net.Listener) {
  self.Transport = &http.Transport{}
  http.Serve(l, self)
}
