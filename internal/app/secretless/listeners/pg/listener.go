package pg

import (
	"fmt"
	"net"
	"strconv"

	"github.com/go-ozzo/ozzo-validation"

	// TODO: Ideally this protocol-specific import shouldn't be needed
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/pg/protocol"
	"github.com/cyberark/secretless-broker/internal/pkg/util"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
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
		if !h.HasCredential("address") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'address'")
		}
		if !h.HasCredential("username") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'username'")
		}
		if !h.HasCredential("password") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'password'")
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
			handlerOptions := plugin_v1.HandlerOptions{
				HandlerConfig:    l.HandlerConfigs[0],
				ClientConnection: client,
				EventNotifier:    l.EventNotifier,
				ShutdownNotifier: func(handler plugin_v1.Handler) {
					l.RemoveHandler(handler)
				},
				Resolver:         l.Resolver,
			}

			handler := l.RunHandlerFunc("pg", handlerOptions)
			l.AddHandler(handler)
		} else {
			pgError := protocol.Error{
				Severity: protocol.ErrorSeverityFatal,
				Code:     protocol.ErrorCodeInternalError,
				Message:  fmt.Sprintf("No handler found for listener %s", l.Config.Name),
			}
			client.Write(pgError.GetMessage())
		}
	}
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "pg"
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	listener :=  &Listener{}
	listener.BaseListener = plugin_v1.NewBaseListener(options, listener)

	return listener
}
