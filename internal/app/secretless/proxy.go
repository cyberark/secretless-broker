package secretless

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
)

// Proxy is the main struct of Secretless.
type Proxy struct {
	Config            config.Config
	EventNotifier     plugin_v1.EventNotifier
	ListenerFactories map[string]func(plugin_v1.ListenerOptions) plugin_v1.Listener
}

// Listen runs the listen loop for a specific Listener.
func (p *Proxy) Listen(listenerConfig config.Listener, wg sync.WaitGroup) {
	var netListener net.Listener
	var err error

	if listenerConfig.Address != "" {
		netListener, err = net.Listen("tcp", listenerConfig.Address)
	} else {
		netListener, err = net.Listen("unix", listenerConfig.Socket)

		// https://stackoverflow.com/questions/16681944/how-to-reliably-unlink-a-unix-domain-socket-in-go-programming-language
		// Handle common process-killing signals so we can gracefully shut down:
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func(c chan os.Signal) {
			// Wait for a SIGINT or SIGKILL:
			sig := <-c
			log.Printf("Caught signal %s: shutting down.", sig)
			// Stop listening (and unlink the socket if unix type):
			netListener.Close()
			// And we're done:
			os.Exit(0)
		}(sigc)
	}
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s listener '%s' listening at: %s",
		listenerConfig.Protocol,
		listenerConfig.Name,
		netListener.Addr())

	options := plugin_v1.ListenerOptions{
		EventNotifier:  p.EventNotifier,
		ListenerConfig: listenerConfig,
		HandlerConfigs: listenerConfig.SelectHandlers(p.Config.Handlers),
		NetListener:    netListener,
	}

	// Ensure that we have this listener
	if _, ok := p.ListenerFactories[listenerConfig.Protocol]; !ok {
		log.Panicf("Unrecognized protocol '%s' on listener '%s'",
			listenerConfig.Protocol, listenerConfig.Name)
	}

	listener := p.ListenerFactories[listenerConfig.Protocol](options)

	err = listener.Validate()
	if err != nil {
		log.Fatalf("Listener '%s' is invalid : %s", listenerConfig.Name, err)
	}

	p.EventNotifier.CreateListener(listener)

	go func() {
		defer wg.Done()
		listener.Listen()
	}()
}

// Run is the main entrypoint to the secretless program.
func (p *Proxy) Run() {
	var wg sync.WaitGroup
	wg.Add(len(p.Config.Listeners))
	for _, config := range p.Config.Listeners {
		p.Listen(config, wg)
	}
	wg.Wait()
}
