package ssh

import (
	"io"
	"log"
	"net"

	"github.com/gliderlabs/ssh"

	"github.com/kgilpin/secretless/internal/pkg/provider"
	"github.com/kgilpin/secretless/pkg/secretless/config"
)

func alwaysPasswordHandler(ctx ssh.Context, password string) bool {
	return true
}

func alwaysPublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}

type Listener struct {
	Config    config.Listener
	Handlers  []config.Handler
	Providers []provider.Provider
	Listener  net.Listener
}

func (self *Listener) Listen() {
	server := &ssh.Server{
		PasswordHandler:  alwaysPasswordHandler,
		PublicKeyHandler: alwaysPublicKeyHandler,
	}

	ssh.Handle(func(s ssh.Session) {
		// Serve the first Handler which is attached to this listener
		var selectedHandler *config.Handler
		for _, handler := range self.Handlers {
			listener := handler.Listener
			if listener == "" {
				listener = handler.Name
			}

			if listener == self.Config.Name {
				selectedHandler = &handler
				break
			}
		}

		if selectedHandler != nil {
			handler := &Handler{Providers: self.Providers, Config: *selectedHandler, Session: s}
			handler.Run()
		} else {
			io.WriteString(s, "No SSH handler is available for this connection!\n")
		}

	})

	log.Fatal(server.Serve(self.Listener))
}
