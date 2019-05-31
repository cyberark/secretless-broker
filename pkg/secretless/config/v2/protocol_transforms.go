/*

These are protocol specific transforms of the configuration.  In the future,
these will be the responsibility of individual service handlers.  We're pulling
them out into their own functions both to clarify what's happening conceptually,
and to prepare for this future refactoring.

*/
package v2

import (
	"fmt"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// v1Service exists purely for clarity, since the concept of a service exists
// implicitly in v1 config, but not anywhere in code.  The combination of a
// Listener and Handler _is_ a v1 service.
type v1Service struct {
	Listener *config.Listener
	Handler *config.Handler
}

func (v1Svc *v1Service) applyProtocolConfig(cfgBytes []byte) error {

	switch v1Svc.Listener.Protocol {
	case "http":
		if err := v1Svc.transformHTTP(cfgBytes); err != nil {
			return err
		}
	}
	return nil
}

func (v1Svc *v1Service) transformHTTP(cfgBytes []byte) error {
	if len(cfgBytes) == 0 {
		return fmt.Errorf("empty http config")
	}

	httpCfg, err := NewHTTPConfig(cfgBytes)
	if err != nil {
		return err
	}

	v1Svc.Handler.Match = httpCfg.AuthenticateURLsMatching

	// TODO: it's funny that this field was only for http, as well as the other
	//  fields we've found this to be true for as well
	v1Svc.Handler.Type = httpCfg.AuthenticationStrategy

	return nil
}
