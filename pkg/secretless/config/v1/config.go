package v1

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-ozzo/ozzo-validation"

	"gopkg.in/yaml.v2"
)

// This represents not the value of a "secret," but the abstract concept of
// "a secret stored somewhere".
//
// Note that "Name" will by the key that maps to this secret's actual value
// in the map[string][]byte when the "StoredSecret" itself is looked up by a
// Resolver.
//
type StoredSecret struct {
	// How client code will refer to the secret's actual value at runtime.
	// Specifically, the key to the secret's value in the map[string][]byte
	// returned by a Resolver.
	Name string
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
	for _, c := range h.Credentials {
		if c.Name == credentialName {
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

// SelectHandlers selects the Handlers that are configured
// to use this Listener.
func (l Listener) SelectHandlers(handlers []Handler) []Handler {
	var result []Handler
	for _, h := range handlers {
		if h.ListenerName == l.Name {
			result = append(result, h)
		}
	}
	return result
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() (err error) {
	err = validation.ValidateStruct(&l,
		validation.Field(&l.Name, validation.Required),
	)

	return
}

type handlerHasListener struct {
	listenerNames map[string]Listener
}

func (hhl handlerHasListener) Validate(value interface{}) error {
	hs := value.([]Handler)
	errors := validation.Errors{}
	for i, h := range hs {
		_, ok := hhl.listenerNames[h.ListenerName]
		if !ok {
			errors[strconv.Itoa(i)] = fmt.Errorf("has no associated listener")
		}
	}
	return errors.Filter()
}

type addressOrSocket struct {
}

func (hhl addressOrSocket) Validate(value interface{}) error {
	ls := value.([]Listener)
	errors := validation.Errors{}
	for i, l := range ls {
		if l.Address == "" && l.Socket == "" {
			errors[strconv.Itoa(i)] = fmt.Errorf("must have an Address or Socket")
		}
	}
	return errors.Filter()
}

func (c Config) String() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(out)
}

// Validate verifies the completeness and correctness of the Config.
func (c Config) Validate() (err error) {
	listenerNames := make(map[string]Listener)
	for _, l := range c.Listeners {
		listenerNames[l.Name] = l
	}

	hasListener := handlerHasListener{listenerNames: listenerNames}

	err = validation.ValidateStruct(&c,
		validation.Field(&c.Handlers, validation.Required, hasListener),
		validation.Field(&c.Listeners, validation.Required, addressOrSocket{}),
		validation.Field(&c.Handlers),
		validation.Field(&c.Listeners),
	)

	return
}
