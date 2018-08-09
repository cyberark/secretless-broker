package example

import (
	"fmt"
	"net"
	"strconv"
	"log"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/cyberark/secretless-broker/internal/pkg/util"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// Listener listens for and handles new connections.
type Listener struct {
	EventNotifier  plugin_v1.EventNotifier
	HandlerConfigs []config.Handler
	NetListener    net.Listener
	Resolver       plugin_v1.Resolver
	Config         config.Listener
	RunHandlerFunc func(id string, options plugin_v1.HandlerOptions)  plugin_v1.Handler
}

// HandlerHasCredentials validates that a handler has all necessary credentials.
type handlerHasCredentials struct {
}

// Validate checks that a handler has all necessary credentials.
func (hhc handlerHasCredentials) Validate(value interface{}) error {
	hs := value.([]config.Handler)
	errors := validation.Errors{}
	for i, h := range hs {
		if !h.HasCredential("host") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'host'")
		}
		if !h.HasCredential("port") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'port'")
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

// Listen listens on the port or socket and attaches new connections to the handler.
func (l *Listener) Listen() {
	for {
		var client net.Conn
		var err error
		if client, err = util.Accept(l); err != nil {
			continue
		}

		// Serve the first Handler which is attached to this listener
		if len(l.HandlerConfigs) > 0 {
			options := plugin_v1.HandlerOptions{
				ClientConnection: client,
				HandlerConfig:    l.HandlerConfigs[0],
				EventNotifier:    l.EventNotifier,
				Resolver:         l.Resolver,
				ShutdownNotifier: func(handler plugin_v1.Handler) {},
			}

			l.RunHandlerFunc("example-handler", options)
		} else {
			client.Write([]byte("Error - no handlers were defined!"))
		}
	}
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "example"
}

// GetConfig implements plugin_v1.Listener
func (l *Listener) GetConfig() config.Listener {
	return l.Config
}

// GetConnections implements plugin_v1.Listener
func (l *Listener) GetConnections() []net.Conn {
	return nil
}

// GetHandlers implements plugin_v1.Listener
func (l *Listener) GetHandlers() []plugin_v1.Handler {
	return nil
}

// GetListener implements plugin_v1.Listener
func (l *Listener) GetListener() net.Listener {
	return l.NetListener
}

// GetNotifier implements plugin_v1.Listener
func (l *Listener) GetNotifier() plugin_v1.EventNotifier {
	return l.EventNotifier
}

// Shutdown implements plugin_v1.Listener
func (l *Listener) Shutdown() error {
	log.Printf("Shutting down example listener's handlers...")

	for _, handler := range l.GetHandlers() {
		handler.Shutdown()
	}

	return l.NetListener.Close()
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	return &Listener{
		EventNotifier:  options.EventNotifier,
		HandlerConfigs: options.HandlerConfigs,
		NetListener:    options.NetListener,
		Resolver:       options.Resolver,
		Config:         options.ListenerConfig,
		RunHandlerFunc: options.RunHandlerFunc,
	}
}
