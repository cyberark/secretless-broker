package interpreterenv

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/containous/yaegi/interp"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
)

// Provider provides the ID as the value.
type Provider struct {
	Name string
}

// ProviderFactory constructs a literal value Provider.
// No configuration or credentials are required.
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	log.Println("Started literalenv provider")
	return &Provider{
		Name: options.Name,
	}, nil
}

// GetName returns the name of the provider
func (p *Provider) GetName() string {
	return p.Name
}

// GetValue returns the id.
func (p *Provider) GetValue(id string) ([]byte, error) {
	providerPath := "./resolver.go"

	log.Printf("Getting variable value of %s from literalenv...", id)

	log.Printf("Reading source file %s...", providerPath)
	file, err := os.Open(providerPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	resolverText, err := ioutil.ReadAll(file)

	log.Println("Provider file read")

	log.Println("Creating interpreter...")
	interpreter := interp.New(interp.Options{})

	log.Println("Evaluating source code...")
	_, err = interpreter.Eval(string(resolverText))
	if err != nil {
		panic(err)
	}

	log.Println("Finding resolve function...")
	v, err := interpreter.Eval("Resolve")
	if err != nil {
		panic(err)
	}

	resolve := v.Interface().(func(string) string)

	log.Println("Invoking resolve function...")
	varValue := resolve(id)

	log.Printf("Got variable value of %s from literalenv: %s", id, varValue)
	return []byte(varValue), nil
}
