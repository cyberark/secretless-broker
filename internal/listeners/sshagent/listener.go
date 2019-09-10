package sshagent

import (
	"fmt"
	"log"

	"github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/ssh/agent"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/util"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// Listener accepts ssh-agent connections and delegates them to the Handler.
type Listener struct {
	plugin_v1.BaseListener
}

// serviceHasCredentials validates that a service has all necessary credentials.
type serviceHasCredentials struct {
}

// Validate checks that a service has all necessary credentials.
func (hhc serviceHasCredentials) Validate(value interface{}) error {
	s := value.(config_v2.Service)

	var err error
	if !s.HasCredential("rsa") && !s.HasCredential("ecdsa") {
		err = fmt.Errorf("must have credential 'rsa' or 'ecdsa'")
	}
	return err
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Config, validation.Required),
		validation.Field(&l.Config, serviceHasCredentials{}),
	)
}

// Listen listens on the ssh-agent socket and attaches new connections to the handler.
func (l *Listener) Listen() {
	keyring := agent.NewKeyring()

	handlerOptions := plugin_v1.HandlerOptions{
		HandlerConfig: l.Config,
		EventNotifier: l.EventNotifier,
		Resolver:      l.Resolver,
	}

	handler := l.RunHandlerFunc("sshagent", handlerOptions)
	if err := handler.LoadKeys(keyring); err != nil {
		log.Printf("Failed to load ssh-agent handler keys: %s", err)
		return
	}

	for l.IsClosed != true {
		nConn, err := util.Accept(l)
		if err != nil {
			log.Printf("WARN: Failed to accept incoming sshagent connection: %s", err)
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
