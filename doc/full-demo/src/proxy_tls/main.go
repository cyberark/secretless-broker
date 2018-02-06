package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func usage() {
	log.Fatal("usage: env SSL_CERT_FILE=<path> SSL_KEY_FILE=<path> proxy_tls <host-and-port>")
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")
	hostAndPort := os.Args[1]
	if certFile == "" || keyFile == "" || hostAndPort == "" {
		usage()
	}

	var url *url.URL
	var err error

	if url, err = url.Parse(fmt.Sprintf("http://%s/", hostAndPort)); err != nil {
		log.Fatalf("Error parsing url '%s' : %s", *url, err)
	}

	log.Printf("Starting myapp_tls on :443")

	proxy := httputil.NewSingleHostReverseProxy(url)
	log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, proxy))
}
