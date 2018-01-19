package secretless

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kgilpin/secretless/internal/app/secretless/http"
	"github.com/kgilpin/secretless/internal/app/secretless/pg"
	"github.com/kgilpin/secretless/internal/app/secretless/ssh"
	"github.com/kgilpin/secretless/internal/app/secretless/sshagent"
	"github.com/kgilpin/secretless/internal/app/secretless/variable"
	"github.com/kgilpin/secretless/internal/pkg/provider"
	"github.com/kgilpin/secretless/pkg/secretless/config"
)

// Listener is an interface for listening in an abstract way.
type Listener interface {
	Listen()
}

// Proxy is the main struct of Secretless.
type Proxy struct {
	Config    config.Config
	Providers []provider.Provider
}

// Listen runs the listen loop for a specific Listener.
func (p *Proxy) Listen(listenerConfig config.Listener, wg sync.WaitGroup) {
	var l net.Listener
	var err error

	if listenerConfig.Address != "" {
		l, err = net.Listen("tcp", listenerConfig.Address)
	} else {
		l, err = net.Listen("unix", listenerConfig.Socket)

		// https://stackoverflow.com/questions/16681944/how-to-reliably-unlink-a-unix-domain-socket-in-go-programming-language
		// Handle common process-killing signals so we can gracefully shut down:
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func(c chan os.Signal) {
			// Wait for a SIGINT or SIGKILL:
			sig := <-c
			log.Printf("Caught signal %s: shutting down.", sig)
			// Stop listening (and unlink the socket if unix type):
			l.Close()
			// And we're done:
			os.Exit(0)
		}(sigc)
	}
	if err == nil {
		log.Printf("%s listener '%s' listening at: %s", listenerConfig.Protocol, listenerConfig.Name, l.Addr())

		protocol := listenerConfig.Protocol
		if protocol == "" {
			protocol = listenerConfig.Name
		}

		var listener Listener
		switch protocol {
		case "pg":
			listener = &pg.Listener{Config: listenerConfig, Listener: l, Providers: p.Providers, Handlers: p.Config.Handlers}
		case "http":
			listener = &http.Listener{Config: listenerConfig, Listener: l, Providers: p.Providers, Handlers: p.Config.Handlers}
		case "ssh":
			listener = &ssh.Listener{Config: listenerConfig, Listener: l, Providers: p.Providers, Handlers: p.Config.Handlers}
		case "ssh-agent":
			listener = &sshagent.Listener{Config: listenerConfig, Listener: l, Providers: p.Providers, Handlers: p.Config.Handlers}
		default:
			panic(fmt.Sprintf("Unrecognized protocol '%s' on listener '%s'", protocol, listenerConfig.Name))
		}
		go func() {
			defer wg.Done()
			listener.Listen()
		}()
	} else {
		log.Fatal(err)
	}
}

// LoadProvider loads a provider from its configuration.
func LoadProvider(providerConfig config.Provider) (provider.Provider, error) {
	pt := providerConfig.Type
	if pt == "" {
		pt = providerConfig.Name
	}

	// TODO: at this time, providers can't load configuration or credentials from each other
	configuration, err := variable.Resolve([]provider.Provider{}, providerConfig.Configuration)
	if err != nil {
		return nil, err
	}
	credentials, err := variable.Resolve([]provider.Provider{}, providerConfig.Credentials)
	if err != nil {
		return nil, err
	}

	switch pt {
	case "environment":
		return provider.NewEnvironmentProvider(providerConfig.Name)
	case "conjur":
		return provider.NewConjurProvider(providerConfig.Name, *configuration, *credentials)
	case "vault":
		return provider.NewVaultProvider(providerConfig.Name, *configuration, *credentials)
	default:
		return nil, fmt.Errorf("Unrecognized provider type '%s'", pt)
	}
}

// Run is the main loop for Secretless.
func (p *Proxy) Run() {
	var err error

	p.Providers = make([]provider.Provider, len(p.Config.Providers))

	for i := range p.Config.Providers {
		p.Providers[i], err = LoadProvider(p.Config.Providers[i])
		if err != nil {
			panic(fmt.Sprintf("Unable to load provider '%s' : %s", p.Config.Providers[i].Name, err.Error()))
		}
		log.Printf("Loaded provider '%s'", p.Providers[i].Name())
	}

	var wg sync.WaitGroup
	wg.Add(len(p.Config.Listeners))
	for _, config := range p.Config.Listeners {
		p.Listen(config, wg)
	}
	wg.Wait()
}
