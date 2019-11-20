package generic

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	validation "github.com/go-ozzo/ozzo-validation"
)

type config struct {
	CredentialPatterns map[string]*regexp.Regexp
	Headers map[string]*template.Template
	ForceSSL bool
}

// validate validates that the given creds satisfy the CredentialValidations of
// the config.
func (c *config) validate(credsByID connector.CredentialValuesByID) error {
	for requiredCred, pattern := range c.CredentialPatterns {
		credVal, ok := credsByID[requiredCred]
		if !ok {
			return fmt.Errorf("missing required credential: %q", requiredCred)
		}
		if !pattern.Match(credVal) {
			return fmt.Errorf(
				"credential %q doesn't match pattern %q", requiredCred, pattern,
			)
		}
	}
	return nil
}

// renderedHeaders returns the config's header templates filled in with the
// given credentialValues.
func (c *config) renderedHeaders(
	credsByID connector.CredentialValuesByID,
) (map[string]string, error) {
	errs := validation.Errors{}
	headers := make(map[string]string)

	// Creds must be strings to work with templates
	credStringsByID := make(map[string]string)
	for credName, credBytes := range credsByID {
		credStringsByID[credName] = string(credBytes)
	}

	for header, tmpl := range c.Headers {
		builder := &strings.Builder{}
		if err := tmpl.Execute(builder, credStringsByID); err != nil {
			errs[header] = fmt.Errorf("couldn't render template: %q", err)
			continue
		}
		headers[header] = builder.String()
	}

	if err := errs.Filter(); err != nil {
		return nil, err
	}

	return headers, nil
}

// newConfig takes a ConfigYAML, validates it, and converts it into a
// generic.config struct -- which is what our application wants to work with.
func newConfig(cfgYAML *ConfigYAML) (*config, error) {
	errs := validation.Errors{}

	cfg := &config{
		CredentialPatterns: make(map[string]*regexp.Regexp),
		Headers: make(map[string]*template.Template),
		ForceSSL: cfgYAML.ForceSSL,
	}

	// Validate and save regexps
	for cred, reStr := range cfgYAML.CredentialValidations {
		re, err := regexp.Compile(reStr)
		if err != nil {
			errs[cred] = fmt.Errorf("invalid regex: %q", err)
			continue
		}
		cfg.CredentialPatterns[cred] = re
	}

	// Validate and save header template strings
	for header, tmplStr := range cfgYAML.Headers {
		tmpl := newHeaderTemplate(header)
		// Ignore pointer to receiver returned by Parse(): it's just "tmpl".
		_, err := tmpl.Parse(tmplStr)
		if err != nil {
			errs[header] = fmt.Errorf("invalid header template: %q", err)
			continue
		}
		cfg.Headers[header] = tmpl
	}

	if err := errs.Filter(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// templateFuncs is a map holding the custom functions available for use within
// a header template string.  We can easily add new functions as needed or
// requested.
var templateFuncs = template.FuncMap{
	"base64": func(str string) string {
		return base64.StdEncoding.EncodeToString([]byte(str))
	},
}

func newHeaderTemplate(name string) *template.Template {
	return template.New(name).Funcs(templateFuncs)
}
