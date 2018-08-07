package mysql

import (
	"fmt"
	"net"
	"strconv"

	// TODO: These errors should be abstracted out ideally
	"github.com/conjurinc/secretless/internal/app/secretless/handlers/mysql/protocol"

	"github.com/conjurinc/secretless/internal/pkg/util"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless/pkg/secretless/plugin/v1"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Listener listens for and handles new connections.
type Listener struct {
	_handlers 	   []plugin_v1.Handler
	Config         config.Listener
	EventNotifier  plugin_v1.EventNotifier
	HandlerConfigs []config.Handler
	NetListener    net.Listener
	Resolver       plugin_v1.Resolver
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
		if !h.HasCredential("host") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'host'")
		}
		if !h.HasCredential("port") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'port'")
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
	l._handlers = make([]plugin_v1.Handler, 0)

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
				Resolver:         l.Resolver,
			}

			handler := l.RunHandlerFunc("mysql", handlerOptions)
			l._handlers = append(l._handlers, handler)

			// TODO: there's a better way to do this
			l.EventNotifier.CreateHandler(handler, client)
		} else {
			mysqlError := protocol.Error{
				Code:     protocol.CRUnknownError,
				SQLSTATE: protocol.ErrorCodeInternalError,
				Message:  fmt.Sprintf("No handler found for listener %s", l.Config.Name),
			}
			client.Write(mysqlError.GetMessage())
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
	return l._handlers
}

// GetConnections implements plugin_v1.Listener
func (l *Listener) GetConnections() []net.Conn {
	return nil
}

// GetNotifier implements plugin_v1.Listener
func (l *Listener) GetNotifier() plugin_v1.EventNotifier {
	return l.EventNotifier
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "mysql"
}

// Shutdown implements plugin_v1.Listener
func (l *Listener) Shutdown() error {
	// TODO: Clean up all handlers
	return l.NetListener.Close()
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	return &Listener{
		Config:         options.ListenerConfig,
		EventNotifier:  options.EventNotifier,
		HandlerConfigs: options.HandlerConfigs,
		NetListener:    options.NetListener,
		Resolver:       options.Resolver,
		RunHandlerFunc: options.RunHandlerFunc,
	}
}
