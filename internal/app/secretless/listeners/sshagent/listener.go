package sshagent

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"golang.org/x/crypto/ssh/agent"

	"github.com/conjurinc/secretless/internal/pkg/util"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/conjurinc/secretless/pkg/secretless/plugin_v1"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Listener accepts ssh-agent connections and delegates them to the Handler.
type Listener struct {
	Config         config.Listener
	EventNotifier  plugin_v1.EventNotifier
	HandlerConfigs []config.Handler
	NetListener    net.Listener
	RunHandlerFunc func(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler
}

// HandlerHasCredentials validates that a handler has all necessary credentials.
type handlerHasCredentials struct {
}

// Validate checks that a handler has all necessary credentials.
func (hhc handlerHasCredentials) Validate(value interface{}) error {
	hs := value.([]config.Handler)
	errors := validation.Errors{}
	for i, h := range hs {
		if !h.HasCredential("rsa") && !h.HasCredential("ecdsa") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'rsa' or 'ecdsa'")
		}
	}
	return errors.Filter()
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.HandlerConfigs, validation.Required),
		validation.Field(&l.HandlerConfigs, handlerHasCredentials{}),
	)
}

// Listen listens on the ssh-agent socket and attaches new connections to the handler.
func (l *Listener) Listen() {
	// Serve the first Handler which is attached to this listener
	if len(l.HandlerConfigs) == 0 {
		log.Panicf("No ssh-agent handler is available")
	}

	selectedHandler := l.HandlerConfigs[0]
	keyring := agent.NewKeyring()

	handlerOptions := plugin_v1.HandlerOptions{
		HandlerConfig: selectedHandler,
		EventNotifier: l.EventNotifier,
	}

	handler := l.RunHandlerFunc("sshagent", handlerOptions)
	if err := handler.LoadKeys(keyring); err != nil {
		log.Printf("Failed to load ssh-agent handler keys: ", err)
		return
	}

	for {
		nConn, err := util.Accept(l)
		if err != nil {
			log.Printf("Failed to accept incoming connection: ", err)
			return
		}

		if err := agent.ServeAgent(keyring, nConn); err != nil {
			log.Printf("Error serving agent : %s", err)
		}
	}
}

// GetConfig implements plugin_v1.Listener
func (l *Listener) GetConfig() config.Listener {
	return l.Config
}

// GetListener implements plugin_v1.Listener
func (l *Listener) GetListener() net.Listener {
	return l.NetListener
}

// GetHandlers implements plugin_v1.Listener
func (l *Listener) GetHandlers() []plugin_v1.Handler {
	return nil
}

// GetConnections implements plugin_v1.Listener
func (l *Listener) GetConnections() []net.Conn {
	return nil
}

// GetNotifier implements plugin_v1.Listener
func (l *Listener) GetNotifier() plugin_v1.EventNotifier {
	return l.EventNotifier
}

func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	return &Listener{
		Config:         options.ListenerConfig,
		EventNotifier:  options.EventNotifier,
		HandlerConfigs: options.HandlerConfigs,
		NetListener:    options.NetListener,
		RunHandlerFunc: options.RunHandlerFunc,
	}
}
