package formatter

import (
	"fmt"

	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/json"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/stdout"
)

var AvailableFormatters = map[string]formatter_api.FormatterConstructor{
	"stdout": stdout.NewFormatter,
	"json":   json.NewFormatter,
}

func GetFormatter(name string, options formatter_api.FormatterOptions) (formatter_api.OutputFormatter, error) {

	formatterConstructor, ok := AvailableFormatters[name]
	if !ok {
		err := fmt.Errorf("ERROR: formatter '%s' not found!", name)
		return nil, err
	}

	formatter, err := formatterConstructor(options)
	if err != nil {
		return nil, err
	}

	return formatter, nil
}
