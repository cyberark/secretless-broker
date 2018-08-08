package example

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cyberark/secretless-broker/internal/pkg/util"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Listener listens for and handles new connections.
type Listener struct {
	plugin_v1.BaseListener
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
				ShutdownNotifier: func(handler plugin_v1.Handler) {
					l.RemoveHandler(handler)
				},
			}

			handler := l.RunHandlerFunc("example-handler", options)
			l.AddHandler(handler)
		} else {
			client.Write([]byte("Error - no handlers were defined!"))
		}
	}
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "example"
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	listener :=  &Listener{}
	listener.BaseListener = plugin_v1.NewBaseListener(options, listener)

	return listener
}
