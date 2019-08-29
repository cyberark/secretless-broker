package v1

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
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

// NewConfig takes the bytes of a file and returns a new Config.
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
		validation.Field(&h.ListenerName, validation.Required),
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

// Validate validates the Listener
func (l Listener) Validate() error {
	// Validations on individual fields of the struct
	fieldErrs := validation.ValidateStruct(
		&l,
		validation.Field(&l.Name, validation.Required),
	)

	// Cast back to validation.Errors so we add to it
	allErrs := validation.Errors{}
	if fieldErrs != nil {
		allErrs = fieldErrs.(validation.Errors)
	}

	// Either Address or Socket must be non-empty
	if l.Address == "" && l.Socket == "" {
		allErrs["AddressOrSocket"] = fmt.Errorf("address or socket is required")
	}

	// Only one of Address or Socket must be non-empty
	if l.Address != "" && l.Socket != "" {
		allErrs["AddressOrSocket"] = fmt.Errorf("only one of address or socket must be provided")
	}

	return allErrs.Filter()
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

	return validation.ValidateStruct(&c,
		validation.Field(&c.Handlers, validation.Required, c.listenerRequired()),
		validation.Field(&c.Listeners, validation.Required, c.handlerRequired()),
		validation.Field(&c.Handlers, validation.Required),
		validation.Field(&c.Listeners, validation.Required),
		validation.Field(&c.Handlers),
		validation.Field(&c.Listeners),
	)
}

// listenerRequired returns a validation.Rule that will ensure every Handler has
// an associated listener.  We cannot define this rule on Handler itself,
// because it needs access to the list of available listeners, which we provide
// here by using a closure.
func (c Config) listenerRequired() validation.Rule {
	availListeners := c.listenersByName()

	// Create a custom validation.RuleFunc
	ruleFunc := func(handlers interface{}) error {
		hs := handlers.([]Handler)
		errors := validation.Errors{}

		for i, h := range hs {
			if _, ok := availListeners[h.ListenerName]; !ok {
				errors[strconv.Itoa(i)] = fmt.Errorf(
					"has no associated listener",
				)
			}
		}

		return errors.Filter()
	}

	return validation.By(ruleFunc)
}

// handlerRequired returns a validation.Rule that will ensure every Listener has
// an associated Handler.  We cannot define this rule on Listener itself,
// because it needs access to the list of available handlers, which we provide
// here by using a closure.
func (c Config) handlerRequired() validation.Rule {
	availHandlers := c.Handlers

	// Create a custom validation.RuleFunc
	ruleFunc := func(listeners interface{}) error {
		errors := validation.Errors{}

		for i, l := range listeners.([]Listener) {
			lHandlers := l.LinkedHandlers(availHandlers)
			if len(lHandlers) == 0 {
				errors[strconv.Itoa(i)] = fmt.Errorf(
					"has no associated handler",
				)
			}
		}

		return errors.Filter()
	}

	return validation.By(ruleFunc)
}

func (c Config) listenersByName() map[string]Listener {
	ret := make(map[string]Listener)
	for _, l := range c.Listeners {
		ret[l.Name] = l
	}
	return ret
}
