package pg

import (
	"fmt"
	"net"

	"github.com/conjurinc/secretless/internal/app/secretless/pg/protocol"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Listener listens for and handles new connections.
type Listener struct {
	Config   config.Listener
	Handlers []config.Handler
	Listener net.Listener
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Handlers, validation.Required),
	)
}

// Listen listens on the port or socket and attaches new connections to the handler.
func (l *Listener) Listen() {
	for {
		var client net.Conn
		var err error
		if client, err = l.Listener.Accept(); err != nil {
			continue
		}

		// Serve the first Handler which is attached to this listener
		if len(l.Handlers) > 0 {
			handler := &Handler{Config: l.Handlers[0], Client: client}
			handler.Run()
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
