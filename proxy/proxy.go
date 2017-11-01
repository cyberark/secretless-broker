package proxy

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/kgilpin/secretless-pg/config"
	"github.com/kgilpin/secretless-pg/proxy/pg"
	"github.com/kgilpin/secretless-pg/proxy/http"
)

type Proxy struct {
	Config config.Config
}

func (self *Proxy) ServePG(config config.ListenerConfig, l net.Listener) {
	for {
		if client, err := l.Accept(); err != nil {
			log.Println(err)
			continue
		} else {
			handler := &pg.PGHandler{Config: config, Client: client}
			go handler.Run()			
		}
	}
}

func (self *Proxy) ServeHTTP(config config.ListenerConfig, l net.Listener) {
	handler := &http.HTTPHandler{Config: config}
	go handler.Run(l)
}

func (self *Proxy) Listen(listenerConfig config.Listener) {
	var proxyListener net.Listener
	var err error

	config := listenerConfig.Config

	if config.Address != "" {
		proxyListener, err = net.Listen("tcp", config.Address)
	} else {
		proxyListener, err = net.Listen("unix", config.Socket)

		// https://stackoverflow.com/questions/16681944/how-to-reliably-unlink-a-unix-domain-socket-in-go-programming-language
		// Handle common process-killing signals so we can gracefully shut down:
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func(c chan os.Signal) {
	    // Wait for a SIGINT or SIGKILL:
	    sig := <-c
	    log.Printf("Caught signal %s: shutting down.", sig)
	    // Stop listening (and unlink the socket if unix type):
	    proxyListener.Close()
	    // And we're done:
	    os.Exit(0)
		}(sigc)		
	}
	if err == nil {
		log.Printf("%s listener '%s' listening at: %s", listenerConfig.Type, listenerConfig.Name, proxyListener.Addr())

		switch listenerConfig.Type {
		case "postgresql": 
			self.ServePG(config, proxyListener)
		case "aws": 
			self.ServeHTTP(config, proxyListener)
		default:
			log.Printf("Unrecognized listener type : %s", listenerConfig.Type)
			return
		}
	} else {
		log.Fatal(err)
	}
}

func (self *Proxy) Run() {
	for _, config := range self.Config.Listeners {
		self.Listen(config)
	}
}
