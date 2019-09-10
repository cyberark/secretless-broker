package pg

import (
	"fmt"
	"log"
	"net"

	"github.com/go-ozzo/ozzo-validation"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/util"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// Listener listens for and handles new connections.
type Listener struct {
	plugin_v1.BaseListener
}

// serviceHasCredentials validates that a service has all necessary credentials.
type serviceHasCredentials struct {
}

// Validate checks that a service has all necessary credentials.
func (svc serviceHasCredentials) Validate(value interface{}) error {
	s := value.(config_v2.Service)
	errors := validation.Errors{}
	if !s.HasCredential("host") && !s.HasCredential("address") {
		errors["host"] = fmt.Errorf("must have credential 'host' or (deprecated) 'address'")
	}

	for _, credential := range [...]string{"username", "password"} {
		if !s.HasCredential(credential) {
			errors[credential] = fmt.Errorf("must have credential '%s'", credential)
		}
	}

	return errors.Filter()
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Config, validation.Required),
		validation.Field(&l.Config, serviceHasCredentials{}),
	)
}

// Listen listens on the port or socket and attaches new connections to the handler.
func (l *Listener) Listen() {
	for l.IsClosed != true {
		var client net.Conn
		var err error
		if client, err = util.Accept(l); err != nil {
			log.Printf("WARN: Failed to accept incoming pg connection: %s", err)
			continue
		}

		// Serve the first Handler which is attached to this listener
		handlerOptions := plugin_v1.HandlerOptions{
			HandlerConfig:    l.Config,
			ClientConnection: client,
			EventNotifier:    l.EventNotifier,
			ShutdownNotifier: func(handler plugin_v1.Handler) {
				l.RemoveHandler(handler)
			},
			Resolver: l.Resolver,
		}

		handler := l.RunHandlerFunc("pg", handlerOptions)
		l.AddHandler(handler)
	}
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "pg"
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	return &Listener{BaseListener: plugin_v1.NewBaseListener(options)}
}
