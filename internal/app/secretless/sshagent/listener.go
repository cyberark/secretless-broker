package sshagent

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"golang.org/x/crypto/ssh/agent"

	"github.com/conjurinc/secretless/pkg/secretless/config"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Listener accepts ssh-agent connections and delegates them to the Handler.
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
		if !h.HasCredential("rsa") && !h.HasCredential("ecdsa") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'rsa' or 'ecdsa'")
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

// Listen listens on the ssh-agent socket and attaches new connections to the handler.
func (l *Listener) Listen() {
	// Serve the first Handler which is attached to this listener
	if len(l.Handlers) == 0 {
		log.Panicf("No ssh-agent handler is available")
	}

	selectedHandler := l.Handlers[0]
	keyring := agent.NewKeyring()

	handler := &Handler{Config: selectedHandler}
	if err := handler.LoadKeys(keyring); err != nil {
		log.Printf("Failed to load ssh-agent handler keys: ", err)
		return
	}

	for {
		nConn, err := l.Listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: ", err)
			return
		}

		if err := agent.ServeAgent(keyring, nConn); err != nil {
			log.Printf("Error serving agent : %s", err)
		}
	}
}
