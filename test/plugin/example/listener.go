package example

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/go-ozzo/ozzo-validation"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/pkg/util"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// Listener listens for and handles new connections.
type Listener struct {
	Config         config_v2.Service
	EventNotifier  plugin_v1.EventNotifier
	NetListener    net.Listener
	Resolver       plugin_v1.Resolver
	RunHandlerFunc func(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler
}

// HandlerHasCredentials validates that a handler has all necessary credentials.
type handlerHasCredentials struct {
}

// Validate checks that a handler has all necessary credentials.
func (hhc handlerHasCredentials) Validate(value interface{}) error {
	hs := value.([]config_v2.Service)
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
		validation.Field(&l.Config, validation.Required),
		validation.Field(&l.Config, handlerHasCredentials{}),
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

		options := plugin_v1.HandlerOptions{
			ClientConnection: client,
			HandlerConfig:    l.Config,
			EventNotifier:    l.EventNotifier,
			Resolver:         l.Resolver,
			ShutdownNotifier: func(handler plugin_v1.Handler) {},
		}

		l.RunHandlerFunc("example-handler", options)

	}
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "example"
}

// GetConfig implements plugin_v1.Listener
func (l *Listener) GetConfig() config_v2.Service {
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
		Config:         options.ServiceConfig,
		EventNotifier:  options.EventNotifier,
		NetListener:    options.NetListener,
		Resolver:       options.Resolver,
		RunHandlerFunc: options.RunHandlerFunc,
	}
}
