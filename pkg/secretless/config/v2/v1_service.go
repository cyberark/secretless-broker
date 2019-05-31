/*

v1Service exists for conceptual clarity.  The concept of a service exists
implicitly in v1 config.Config, but not in the code.  The combination of a
Listener and Handler implicitly represents a service in the v1 code. We're
making it explicit here.

v1Service also hourses protocol specific configuration logic.  In the future,
this logic will be the responsibility of individual v2 services.  We're pulling
them out now into their own functions both to clarify that this is a separate
step of the configuration process -- one specific to each protocol -- and to
prepare for this future refactoring.

*/
package v2

import (
	"fmt"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

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
