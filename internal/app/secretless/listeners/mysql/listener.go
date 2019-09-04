package mysql

import (
	"fmt"
	"log"
	"net"

	"github.com/go-ozzo/ozzo-validation"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/pkg/util"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
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
	h := value.(config_v2.Service)

	var err error
	if !h.HasCredential("host") {
		err = fmt.Errorf("must have credential 'host'")
	}
	if !h.HasCredential("port") {
		err = fmt.Errorf("must have credential 'port'")
	}
	if !h.HasCredential("username") {
		err = fmt.Errorf("must have credential 'username'")
	}
	if !h.HasCredential("password") {
		err = fmt.Errorf("must have credential 'password'")
	}

	return err
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
	for l.IsClosed != true {
		var client net.Conn
		var err error
		if client, err = util.Accept(l); err != nil {
			log.Printf("WARN: Failed to accept incoming mysql connection: %s", err)
			continue
		}

		// Serve the first Handler which is attached to this listener
		handlerOptions := plugin_v1.HandlerOptions{
			HandlerConfig:    l.Config,
			ClientConnection: client,
			EventNotifier:    l.EventNotifier,
			Resolver:         l.Resolver,
			ShutdownNotifier: func(handler plugin_v1.Handler) {
				l.RemoveHandler(handler)
			},
		}

		handler := l.RunHandlerFunc("mysql", handlerOptions)
		l.AddHandler(handler)
	}
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "mysql"
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	return &Listener{BaseListener: plugin_v1.NewBaseListener(options)}
}
