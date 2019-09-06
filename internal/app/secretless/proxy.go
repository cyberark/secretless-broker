package secretless

import (
	"log"
	"net"
	"strings"
	"sync"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/pkg/util"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

const (
	// START is a signal that indicates that the proxy is instantiating
	START = iota
	// RESTART event signals that a reload of configuration is needed
	RESTART
	// SHUTDOWN signals that that we are in the process of program exit
	SHUTDOWN
)

// Proxy is the main struct of Secretless.
type Proxy struct {
	cleanupMutex    sync.Mutex
	runEventChan    chan int
	EventNotifier   plugin_v1.EventNotifier
	Config          config_v2.Config
	Listeners       []plugin_v1.Listener
	Resolver        plugin_v1.Resolver
	RunHandlerFunc  func(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler
	RunListenerFunc func(id string, options plugin_v1.ListenerOptions) plugin_v1.Listener
}

// Listen runs the listen loop for a specific Listener.
func (p *Proxy) Listen(listenerConfig config_v2.Service) plugin_v1.Listener {
	network := "tcp"
	if strings.HasPrefix(listenerConfig.ListenOn, "unix") {
		network = "unix"
	}
	address := strings.TrimPrefix(listenerConfig.ListenOn, network + "://")
	netListener, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s listener '%s' listening at: %s",
		listenerConfig.Connector,
		listenerConfig.Name,
		netListener.Addr())

	options := plugin_v1.ListenerOptions{
		EventNotifier:  p.EventNotifier,
		ServiceConfig: listenerConfig,
		NetListener:    netListener,
		Resolver:       p.Resolver,
		RunHandlerFunc: p.RunHandlerFunc,
	}

	listenerID := listenerConfig.Connector
	// At present, we still need to use the http listener for http service connectors
	if config_v2.IsHTTPConnector(listenerID) {
		listenerID = "http"
	}
	listener := p.RunListenerFunc(listenerID, options)

	err = listener.Validate()
	if err != nil {
		log.Fatalf("Listener '%s' is invalid : %s", listenerConfig.Name, err)
	}

	p.EventNotifier.CreateListener(listener)

	go func() {
		defer listener.Shutdown()
		listener.Listen()
	}()

	return listener
}

// ReloadListeners sends RESTART msg to runEventChan
func (p *Proxy) ReloadListeners() error {
	p.runEventChan <- RESTART

	// TODO: Return any errors we get during reload
	return nil
}

// Shutdown sends SHUTDOWN msg to runEventChan
func (p *Proxy) Shutdown() {
	p.runEventChan <- SHUTDOWN
	p.cleanUpListeners()
}

// Loops through the listeners and shuts them down concurrently
func (p *Proxy) cleanUpListeners() {
	// because cleanUpListeners can be called from different goroutines
	defer p.cleanupMutex.Unlock()
	p.cleanupMutex.Lock()

	var wg sync.WaitGroup

	for _, listener := range p.Listeners {
		// block scoped variable for use in goroutine
		_listener := listener

		log.Printf("Shutting down '%v' listener...", listener.GetName())

		wg.Add(1)
		go func() {
			defer wg.Done()
			_listener.Shutdown()
		}()
	}

	wg.Wait()
}

// Run is the main entrypoint to the secretless program.
// the for-select loop allows for queueing of RESTARTS and only 1 SHUTDOWN
func (p *Proxy) Run() {
	p.runEventChan = make(chan int, 1)
	p.cleanupMutex = sync.Mutex{}

	go func() {
		p.runEventChan <- START
	}()

	// When runEventChan receives message...
	// START, RESTART: runs cleanUpListeners and reloads all listeners
	// SHUTDOWN: proceed to infinite non-busy for-loop
	// default: panic
	for msg := range p.runEventChan {
		switch msg {
		case START, RESTART:
			// Set the health check to show that we are ready
			if msg == START {
				util.SetAppInitializedFlag()
			}

			// Mark app as "live"
			util.SetAppIsLive(false)

			p.cleanUpListeners()
			p.Listeners = make([]plugin_v1.Listener, 0)

			// TODO: Delegate logic of this `if` check to connection managers
			// XXX: App is not marked as live if we are not listening to anything. This
			//      may need clarification later.
			if len(p.Config.Services) < 1 {
				log.Println("Waiting for valid configuration to be provided...")
				break
			}

			log.Println("Starting all listeners and handlers...")
			for _, config := range p.Config.Services {
				listener := p.Listen(*config)
				p.Listeners = append(p.Listeners, listener)
			}

			util.SetAppIsLive(true)

		case SHUTDOWN:
			log.Println("Shutdown requested. Waiting for cleanup...")
			// Block forever until explicit os.Exit and prevent processing
			// of further runEventChan messages
			select {}
		default:
			log.Panic("Proxy#Run should never reach here")
		}
	}
}
