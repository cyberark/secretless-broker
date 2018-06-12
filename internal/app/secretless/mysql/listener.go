package mysql

import (
	"fmt"
	"net"
	"strconv"

	"github.com/conjurinc/secretless/internal/app/secretless/mysql/protocol"
	"github.com/conjurinc/secretless/internal/pkg/util"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/conjurinc/secretless/pkg/secretless/handler"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Listener listens for and handles new connections.
type Listener struct {
	Config   config.Listener
	Handlers []config.Handler
	Listener net.Listener
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
		validation.Field(&l.Handlers, validation.Required),
		validation.Field(&l.Handlers, handlerHasCredentials{}),
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
		if len(l.Handlers) > 0 {
			handler := &Handler{Config: l.Handlers[0], Client: client}
			handler.Run()
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

// GetConfig implements secretless.Listener
func (l *Listener) GetConfig() config.Listener {
	return l.Config
}

// GetListener implements secretless.Listener
func (l *Listener) GetListener() net.Listener {
	return l.Listener
}

// GetHandlers implements secretless.Listener
func (l *Listener) GetHandlers() []handler.Handler_v1 {
	return nil
}

// GetConnections implements secretless.Listener
func (l *Listener) GetConnections() []net.Conn {
	return nil
}
