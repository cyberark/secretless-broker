package generic

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	validation "github.com/go-ozzo/ozzo-validation"
)

type config struct {
	CredentialPatterns map[string]*regexp.Regexp
	Headers            map[string]*template.Template
	OAuth1Secrets      map[string]*template.Template
	QueryParams        map[string]*template.Template
	ForceSSL           bool
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

// renderTemplates returns the config's templates filled in with the
// given credentialValues.
func renderTemplates(
	template map[string]*template.Template,
	credsByID connector.CredentialValuesByID,
) (map[string]string, error) {
	errs := validation.Errors{}
	args := make(map[string]string)

	// Creds must be strings to work with templates
	credStringsByID := make(map[string]string)
	for credName, credBytes := range credsByID {
		credStringsByID[credName] = string(credBytes)
	}

	for arg, tmpl := range template {
		builder := &strings.Builder{}
		if err := tmpl.Execute(builder, credStringsByID); err != nil {
			errs[arg] = fmt.Errorf("couldn't render template: %q", err)
			continue
		}
		args[arg] = builder.String()
	}

	if err := errs.Filter(); err != nil {
		return nil, err
	}

	return args, nil
}

// newConfig takes a ConfigYAML, validates it, and converts it into a
// generic.config struct -- which is what our application wants to work with.
func newConfig(cfgYAML *ConfigYAML) (*config, error) {
	errs := validation.Errors{}

	cfg := &config{
		CredentialPatterns: make(map[string]*regexp.Regexp),
		ForceSSL:           cfgYAML.ForceSSL,
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

	cfg.Headers, errs = stringsToTemplates(cfgYAML.Headers, errs)
	cfg.QueryParams, errs = stringsToTemplates(cfgYAML.QueryParams, errs)
	cfg.OAuth1Secrets, errs = stringsToTemplates(cfgYAML.OAuth1Secrets, errs)

	if err := errs.Filter(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func stringsToTemplates(
	templates map[string]string,
	errs validation.Errors,
) (map[string]*template.Template, validation.Errors) {
	parsedTemplates := make(map[string]*template.Template)
	// Validate and save template strings
	for tmplName, tmplStr := range templates {
		tmpl := newHTTPTemplate(tmplName)
		// Ignore pointer to receiver returned by Parse(): it's just "tmpl".
		_, err := tmpl.Parse(tmplStr)
		if err != nil {
			errs[tmplName] = fmt.Errorf("invalid template: %q", err)
			continue
		}
		parsedTemplates[tmplName] = tmpl
	}
	return parsedTemplates, errs
}

func appendQueryParams(URL url.URL, params map[string]string) string {
	query := url.Values{}
	if len(URL.RawQuery) > 0 {
		query = URL.Query()
	}
	for key, value := range params {
		query.Add(key, value)
	}

	return query.Encode()
}

// templateFuncs is a map holding the custom functions available for use within
// a header template string.  We can easily add new functions as needed or
// requested.
var templateFuncs = template.FuncMap{
	"base64": func(str string) string {
		return base64.StdEncoding.EncodeToString([]byte(str))
	},
}

func newHTTPTemplate(name string) *template.Template {
	return template.New(name).Funcs(templateFuncs)
}
