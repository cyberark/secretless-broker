package secretless

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/conjurinc/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless-broker/pkg/secretless/plugin/v1"
)

// Proxy is the main struct of Secretless.
type Proxy struct {
	Config            config.Config
	EventNotifier     plugin_v1.EventNotifier
	Listeners         []plugin_v1.Listener
	ListenerWaitGroup sync.WaitGroup
	Resolver          plugin_v1.Resolver
	RunListenerFunc   func(id string, options plugin_v1.ListenerOptions) plugin_v1.Listener
	RunHandlerFunc    func(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler
}

// Listen runs the listen loop for a specific Listener.
func (p *Proxy) Listen(listenerConfig config.Listener, wg sync.WaitGroup) plugin_v1.Listener {
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
		Resolver:       p.Resolver,
		RunHandlerFunc: p.RunHandlerFunc,
	}

	listener := p.RunListenerFunc(listenerConfig.Protocol, options)

	err = listener.Validate()
	if err != nil {
		log.Fatalf("Listener '%s' is invalid : %s", listenerConfig.Name, err)
	}

	p.EventNotifier.CreateListener(listener)

	go func() {
		defer wg.Done()
		listener.Listen()
	}()

	return listener
}

// ReloadListeners will loop through the listeners and shut them down
// As each listener is shut down the WaitGroup is decremented, and once the
// counter is zero the Proxy.Run loop will complete and restart, reloading all
// of the listeners.
func (p *Proxy) ReloadListeners() error {
	if p.Listeners == nil || len(p.Listeners) == 0 {
		log.Println("WARN: No listeners to reload!")
		return nil
	}

	for _, listener := range p.Listeners {
		log.Printf("Shutting down '%v' listener...", listener.GetName())
		listener.Shutdown()
		p.ListenerWaitGroup.Done()
	}

	// TODO: Return any errors we get during reload
	return nil
}

// Run is the main entrypoint to the secretless program.
func (p *Proxy) Run() {
	p.ListenerWaitGroup = sync.WaitGroup{}
	// We loop until we get an exit signal (in which case we exit program)
	for {
		// TODO: Delegate logic of this `if` check to connection managers
		if len(p.Config.Listeners) < 1 {
			log.Fatalln("ERROR! No listeners specified in config!")
		}

		p.Listeners = make([]plugin_v1.Listener, 0)
		log.Println("Starting all listeners and handlers...")
		p.ListenerWaitGroup.Add(len(p.Config.Listeners))
		for _, config := range p.Config.Listeners {
			listener := p.Listen(config, p.ListenerWaitGroup)
			p.Listeners = append(p.Listeners, listener)
		}
		p.ListenerWaitGroup.Wait()
	}
}
