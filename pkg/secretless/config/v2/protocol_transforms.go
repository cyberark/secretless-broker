/*

These are protocol specific transforms of the configuration.  In the future, these will be the
responsibility of individual service handlers.  We're pulling them out into
their own functions both to clarify what's happening conceptually, and to
prepare for this future refactoring.

*/
package v2

import (
	"fmt"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

type protocolTransform func(bytes []byte, listener *config.Listener, handler *config.Handler) ( error)

var protocolTransforms = map[string]protocolTransform{
	"http": func(cfgBytes []byte, listener *config.Listener, handler *config.Handler) error {
		if len(cfgBytes) == 0 {
			return fmt.Errorf("http config: nil")
		}

		httpCfg, err := NewHTTPConfig(cfgBytes)
		if err != nil {
			return err
		}

		handler.Match = httpCfg.AuthenticateURLsMatching
		// TODO: it's funny that this field was only for http, as well as the other fields we've found this to be true for as well
		handler.Type = httpCfg.AuthenticationStrategy

		return nil
	},
}

func applyProtocolTransform(protocol string, configBytes []byte, listener *config.Listener, handler *config.Handler) error {

	if lhTransform, ok := protocolTransforms[protocol]; ok {
		return lhTransform(configBytes, listener, handler)
	}

	return nil
}
