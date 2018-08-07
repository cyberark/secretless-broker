package secretless

import (
	"log"
	"net"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
	"sync"
)

const (
	RESTART = iota
	SHUTDOWN
	KILL
)

// Proxy is the main struct of Secretless.
type Proxy struct {
	Config            config.Config
	EventNotifier     plugin_v1.EventNotifier
	Listeners         []plugin_v1.Listener
	_runCh            chan int
	_cleanUpMux       sync.Mutex
	Resolver          plugin_v1.Resolver
	RunListenerFunc   func(id string, options plugin_v1.ListenerOptions) plugin_v1.Listener
	RunHandlerFunc    func(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler
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
		defer shutDownListener(listener)
		listener.Listen()
	}()

	return listener
}

// ReloadListeners sends RESTART msg to _runCh
func (p *Proxy) ReloadListeners() error {
	p._runCh <- RESTART

	// TODO: Return any errors we get during reload
	return nil
}

// Shutdown sends SHUTDOWN msg to _runCh
func (p *Proxy) Shutdown() {
	p._runCh <- SHUTDOWN
	p.cleanUpListeners()
}

// Loops through the listeners and shuts them down concurrently
func (p *Proxy) cleanUpListeners() {
	var wg sync.WaitGroup

	for _, _listener := range p.Listeners {
		listener := _listener

		log.Printf("Shutting down '%v' listener...", listener.GetName())

		wg.Add(1)
		go func() {
			defer wg.Done()
			shutDownListener(listener)
		}()
	}

	wg.Wait()
}

// Run is the main entrypoint to the secretless program.
// the for-select loop allows for queueing of RESTARTS and only 1 SHUTDOWN
func (p *Proxy) Run() {
	p._runCh = make(chan int, 1)
	p._cleanUpMux = sync.Mutex{}

	go func() {
		p._runCh <- RESTART
	}()

	// when _runCh receives message...
	// RESTART: runs cleanUpListeners and reloads all listeners
	// SHUTDOWN: for-select turns to infinite non-busy loop
	// default: panic
	for {
		select {
		case msg := <-p._runCh:
			switch msg {
			case RESTART:
				p.cleanUpListeners()
				// TODO: Delegate logic of this `if` check to connection managers
				if len(p.Config.Listeners) < 1 {
					log.Fatalln("ERROR! No listeners specified in config!")
				}

				p.Listeners = make([]plugin_v1.Listener, 0)
				log.Println("Starting all listeners and handlers...")
				for _, config := range p.Config.Listeners {
					listener := p.Listen(config)
					p.Listeners = append(p.Listeners, listener)
				}
			case SHUTDOWN:
				// non-busy for-select loops forever until explicit os.Exit
				p._runCh = nil
			default:
				log.Panic("Proxy#Run should never reach here")
			}
		}
	}
}

// util
func shutDownListener(listener plugin_v1.Listener) {
	log.Printf("Shutting down '%v' listener's handlers...", listener.GetName())

	var wg sync.WaitGroup

	for _, _handler := range listener.GetHandlers() {
		handler := _handler

		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.Shutdown()
		}()
	}

	wg.Wait()

	listener.Shutdown()
}
