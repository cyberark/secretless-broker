package http

import (
  "io"
  "log"
  "net"
  "net/http"

  "github.com/kgilpin/secretless/config"
)

type Authenticator interface {
  Authenticate(*http.Request) error
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

// Attribution: https://github.com/elazarl/goproxy/blob/de25c6ed252fdc01e23dae49d6a86742bd790b12/proxy.go#L74
func removeProxyHeaders(r *http.Request) {
  r.RequestURI = "" // this must be reset when serving a request with the client
  log.Printf("Sending request %v %v", r.Method, r.URL.String())
  // If no Accept-Encoding header exists, Transport will add the headers it can accept
  // and would wrap the response body with the relevant reader.
  r.Header.Del("Accept-Encoding")
  // curl can add that, see
  // https://jdebp.eu./FGA/web-proxy-connection-header.html
  r.Header.Del("Proxy-Connection")
  r.Header.Del("Proxy-Authenticate")
  r.Header.Del("Proxy-Authorization")
  // Connection, Authenticate and Authorization are single hop Header:
  // http://www.w3.org/Protocols/rfc2616/rfc2616.txt
  // 14.10 Connection
  //   The Connection general-header field allows the sender to specify
  //   options that are desired for that particular connection and MUST NOT
  //   be communicated by proxies over further connections.
  r.Header.Del("Connection")
  r.Header.Del("Authenticate")
  r.Header.Del("Authorization")
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
    log.Printf("Request headers: %v", r.Header)

    if !r.URL.IsAbs() {
      http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", 500)
      return
    }

    removeProxyHeaders(r)

    if err = self.Authenticator.Authenticate(r); err != nil {
      http.Error(w, err.Error(), 500)
      return
    }

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
