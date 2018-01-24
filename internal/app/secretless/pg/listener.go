package pg

import (
	"fmt"
	"net"

	"github.com/conjurinc/secretless/internal/app/secretless/pg/protocol"
	"github.com/conjurinc/secretless/internal/pkg/provider"
	"github.com/conjurinc/secretless/pkg/secretless/config"
)

// Listener listens for and handles new connections.
type Listener struct {
	Config    config.Listener
	Handlers  []config.Handler
	Providers []provider.Provider
	Listener  net.Listener
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
		var selectedHandler *config.Handler
		for _, handler := range l.Handlers {
			listener := handler.Listener
			if listener == "" {
				listener = handler.Name
			}

			if listener == l.Config.Name {
				selectedHandler = &handler
				break
			}
		}

		if selectedHandler != nil {
			handler := &Handler{Providers: l.Providers, Config: *selectedHandler, Client: client}
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
