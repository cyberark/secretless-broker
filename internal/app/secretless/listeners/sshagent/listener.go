package sshagent

import (
	"fmt"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/sshagent"
	"log"
	"strconv"

	"github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/ssh/agent"

	"github.com/cyberark/secretless-broker/internal/pkg/util"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// Listener accepts ssh-agent connections and delegates them to the Handler.
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
		if !h.HasCredential("rsa") && !h.HasCredential("ecdsa") {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have credential 'rsa' or 'ecdsa'")
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

// Listen listens on the ssh-agent socket and attaches new connections to the handler.
func (l *Listener) Listen() {
	// Serve the first Handler which is attached to this listener
	if len(l.HandlerConfigs) == 0 {
		log.Panicf("No ssh-agent handler is available")
	}

	selectedHandler := l.HandlerConfigs[0]
	keyring := agent.NewKeyring()

	handlerOptions := plugin_v1.HandlerOptions{
		HandlerConfig: selectedHandler,
		EventNotifier: l.EventNotifier,
		Resolver:      l.Resolver,
	}

	genericHandler := l.RunHandlerFunc("sshagent", handlerOptions)
// NOTE:
// This type coercion is a hack to prevent an even worse hack.  Originally, we had
// a method `LoadKeys` on the Handler interface itself, even though the method was
// only needed for this specific handler.  We decided that keeping the public
// interface clean was far more important, but in order to do that we must do this
// ugly type assertion in this handler.  Eventually, a more comprehensive refactor
// of the Secretless architecture will clean this up.
	handler, ok := genericHandler.(*sshagent.Handler)
	if !ok {
		log.Printf("handler created is not for ssh-agent")
		return
	}

	if err := handler.LoadKeys(keyring); err != nil {
		log.Printf("Failed to load ssh-agent handler keys: ", err)
		return
	}

	for l.IsClosed != true {
		nConn, err := util.Accept(l)
		if err != nil {
			log.Printf("WARN: Failed to accept incoming sshagent connection: ", err)
			return
		}

		if err := agent.ServeAgent(keyring, nConn); err != nil {
			log.Printf("Error serving agent : %s", err)
		}
	}
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "sshagent"
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	return &Listener{BaseListener: plugin_v1.NewBaseListener(options)}
}
