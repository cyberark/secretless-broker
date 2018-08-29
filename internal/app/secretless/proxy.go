package secretless

import (
	"log"
	"net"
	"sync"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
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
	Config          config.Config
	Listeners       []plugin_v1.Listener
	Resolver        plugin_v1.Resolver
	RunHandlerFunc  func(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler
	RunListenerFunc func(id string, options plugin_v1.ListenerOptions) plugin_v1.Listener
}

// Listen runs the listen loop for a specific Listener.
func (p *Proxy) Listen(listenerConfig config.Listener) plugin_v1.Listener {
	var netListener net.Listener
	var err error

	if listenerConfig.Address != "" {
		netListener, err = net.Listen("tcp", listenerConfig.Address)
	} else {
		netListener, err = net.Listen("unix", listenerConfig.Socket)
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
			p.cleanUpListeners()
			p.Listeners = make([]plugin_v1.Listener, 0)

			// TODO: Delegate logic of this `if` check to connection managers
			if len(p.Config.Listeners) < 1 {
				log.Println("WARN! No listeners specified in config (wait loop)!")
				break
			}

			log.Println("Starting all listeners and handlers...")
			for _, config := range p.Config.Listeners {
				listener := p.Listen(config)
				p.Listeners = append(p.Listeners, listener)
			}
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
