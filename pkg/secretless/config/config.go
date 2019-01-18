package config

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/go-ozzo/ozzo-validation"

	yaml "gopkg.in/yaml.v2"

	crd_api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
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

func parseConfigFile(buffer []byte, config *Config) (err error) {
	if err = yaml.UnmarshalStrict(buffer, &config); err != nil {
		return err
	}
	return
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

// LoadFromFile loads a configuration file into a Config object.
func LoadFromFile(fileName string) (config Config, err error) {
	var buffer []byte
	if buffer, err = ioutil.ReadFile(fileName); err != nil {
		err = fmt.Errorf("Error reading config file '%s': '%s'", fileName, err)
		return
	}
	return Load(buffer)
}

// LoadFromCRD loads a configuration from a CRD API Configuration object
func LoadFromCRD(crdConfig crd_api_v1.Configuration) (config Config, err error) {
	var specData []byte
	if specData, err = yaml.Marshal(crdConfig.Spec); err != nil {
		return
	}

	if config, err = Load(specData); err != nil {
		return
	}

	return config, nil
}

// Load parses a YAML string into a Config object.
func Load(data []byte) (config Config, err error) {
	if err = parseConfigFile(data, &config); err != nil {
		err = fmt.Errorf("Unable to parse configuration: '%s'", err)
		return
	}

	for i := range config.Listeners {
		l := &config.Listeners[i]
		if l.Protocol == "" {
			l.Protocol = l.Name
		}
	}

	for i := range config.Handlers {
		h := &config.Handlers[i]
		if h.Type == "" {
			h.Type = h.Name
		}
		if h.ListenerName == "" {
			h.ListenerName = h.Name
		}

		h.Patterns = make([]*regexp.Regexp, len(h.Match))
		for i, matchPattern := range h.Match {
			pattern, err := regexp.Compile(matchPattern)
			if err != nil {
				panic(err.Error())
			} else {
				h.Patterns[i] = pattern
			}
		}
	}

	if err = config.Validate(); err != nil {
		err = fmt.Errorf("Configuration is not valid: %s", err)
		return
	}

	return
}
