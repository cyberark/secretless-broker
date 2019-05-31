package v2

import (
	"fmt"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

type listenerHandlerTransform func(bytes []byte, listener *config.Listener, handler *config.Handler) ( error)

var listenerHandlerTransforms = map[string]listenerHandlerTransform{
	"http": func(cfgBytes []byte, listener *config.Listener, handler *config.Handler) error {
		if len(cfgBytes) == 0 {
			return fmt.Errorf("http config: nil")
		}

		hTTPConfig, err := NewHTTPConfig(cfgBytes)
		if err != nil {
			return err
		}

		handler.Match = hTTPConfig.AuthenticateURLsMatching
		// TODO: it's funny that this field was only for http, as well as the other fields we've found this to be true for as well
		handler.Type = hTTPConfig.AuthenticationStrategy

		return nil
	},
}

func transformListenerHandler(protocol string, configBytes []byte, listener *config.Listener, handler *config.Handler) error {
	if lhTransform, ok := listenerHandlerTransforms[protocol]; ok {
		return lhTransform(configBytes, listener, handler)
	}

	return nil
}
