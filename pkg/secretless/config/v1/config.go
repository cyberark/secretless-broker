package v1

import (
	"fmt"
	"github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
	"regexp"
)

// StoredSecret represents not the value of a "secret," but the abstract concept
// of "a secret stored somewhere".
//
// Note that "Name" will be the key that maps to this secret's actual value in
// the map[string][]byte when the "StoredSecret" itself is looked up by a
// Resolver.
type StoredSecret struct {
	// How client code will refer to the secret's actual value at runtime.
	// Specifically, the key to the secret's value in the map[string][]byte
	// returned by a Resolver.
	Name     string
	Provider string
	// The identifier within the context of a Provider.  Ie, how a provider
	// refers to this secret.  Eg, a database primary key.
	ID string
}

// Listener listens on a port on socket for inbound connections, which are
// handed off to Handlers.
type Listener struct {
	Address     string
	CACertFiles []string `yaml:"caCertFiles"`
	Debug       bool
	Name        string
	Protocol    string
	Socket      string
}

// Handler processes an inbound message and connects to a specified backend
// using Credentials which it fetches from a provider.
type Handler struct {
	Name         string
	Type         string
	ListenerName string `yaml:"listener"`
	Debug        bool
	Match        []string         `yaml:"match"`
	Patterns     []*regexp.Regexp `yaml:"-"`
	Credentials  []StoredSecret
}

// Config is the main configuration structure for Secretless.
// It lists and configures the protocol listeners and handlers.
type Config struct {
	Listeners []Listener
	Handlers  []Handler
}

// HasCredential indicates whether a Handler has the specified credential.
func (h Handler) HasCredential(credentialName string) bool {
	for _, credential := range h.Credentials {
		if credential.Name == credentialName {
			return true
		}
	}
	return false
}

func NewConfig(buffer []byte) (*Config, error) {
	config := &Config{}
	if err := yaml.Unmarshal(buffer, config); err != nil {
		return nil, err
	}
	return config, nil
}

// Validate verifies the completeness and correctness of the Handler.
func (h Handler) Validate() (err error) {
	err = validation.ValidateStruct(&h,
		validation.Field(&h.Name, validation.Required),
	)

	return
}

// LinkedHandlers filters the handlers passed to it, returning only those
// attached to this Listener
func (l Listener) LinkedHandlers(handlers []Handler) []Handler {
	var result []Handler
	for _, h := range handlers {
		if h.ListenerName == l.Name {
			result = append(result, h)
		}
	}
	return result
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() error {
	return validation.ValidateStruct(
		&l,
		validation.Field(&l.Name, validation.Required),
	)
}

// addressOrSocketRequiredRule is validation.RuleFunc that ensures a Listener
// has either an Address or Socket configured.  This is a helper func for
// Listener.Validate().
func hasAddressOrSocket(v interface{}) error {
	l := v.(Listener)
	if l.Address == "" && l.Socket == "" {
		return fmt.Errorf("either Address or Socket is required")
	}
	return nil
}

func (c Config) String() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(out)
}

// Validate verifies the completeness and correctness of the Config.
func (c Config) Validate() error {

	// Create a rule than ensures every Handler has a Listener
	var listenerNames []interface{}
	for _, l := range c.Listeners {
		listenerNames = append(listenerNames, l.Name)
	}
	listenerRequired := validation.In(listenerNames...)

	return validation.ValidateStruct(&c,
		validation.Field(&c.Handlers, validation.Required, listenerRequired),
		validation.Field(&c.Listeners, validation.Required, validation.By(hasAddressOrSocket)),
		validation.Field(&c.Handlers),
		validation.Field(&c.Listeners),
	)
}
