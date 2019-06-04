/*

v1Service exists for conceptual clarity.  The concept of a service exists
implicitly in v1.Config, but not in the code.  The combination of a Listener and
Handler implicitly represents a service in the v1 code. We're making it explicit
here.

v1Service also houses protocol specific configuration logic.  In the future,
this logic will be the responsibility of individual v2 services.  We're pulling
them out now into their own functions both to clarify that this is a separate
step of the configuration process -- one specific to each protocol -- and to
prepare for this future refactoring.

*/
package v2

import (
	"fmt"
	"sort"
	"strings"

	config_v1 "github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
)

type v1Service struct {
	Listener *config_v1.Listener
	Handler *config_v1.Handler
}

func newV1Service(v2Svc Service) (ret *v1Service, err error) {

	// Create basic Service

	ret = &v1Service{
		Listener: &config_v1.Listener{
			Name:     v2Svc.Name,
			Protocol: v2Svc.Protocol,
		},
		Handler: &config_v1.Handler{
			Name:         v2Svc.Name,
			ListenerName: v2Svc.Name,
		},
	}

	// Map ListenOn To Address or Socket

	if strings.HasPrefix(v2Svc.ListenOn, "tcp://") {
		ret.Listener.Address = strings.TrimPrefix(v2Svc.ListenOn, "tcp://")
	} else if strings.HasPrefix(v2Svc.ListenOn, "unix://") {
		ret.Listener.Socket = strings.TrimPrefix(v2Svc.ListenOn, "unix://")
	} else {
		errMsg := "listenOn=%q missing prefix from one of tcp:// or unix//"
		return nil, fmt.Errorf(errMsg, v2Svc.ListenOn)
	}

	// Map v2.Credentials to v1.StoredSecret

	credentials := make([]config_v1.StoredSecret, 0)
	for _, cred := range v2Svc.Credentials {
		credentials = append(credentials, config_v1.StoredSecret{
			Name:     cred.Name,
			Provider: cred.From,
			ID:       cred.Get,
		})
	}

	// Sort Credentials

	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].Name < credentials[j].Name
	})

	// Add Credentials to Handler

	ret.Handler.Credentials = credentials

	// Apply protocol specific config

	if err = ret.applyProtocolConfig(v2Svc.ProtocolConfig); err != nil {
		return nil, err
	}

	return ret, nil
}

func (v1Svc *v1Service) applyProtocolConfig(cfgBytes []byte) error {

	switch v1Svc.Listener.Protocol {
	case "http":
		if err := v1Svc.configureHTTP(cfgBytes); err != nil {
			return err
		}
	}
	return nil
}

func (v1Svc *v1Service) configureHTTP(cfgBytes []byte) error {
	if len(cfgBytes) == 0 {
		return fmt.Errorf("empty http config")
	}

	httpCfg, err := newHTTPConfig(cfgBytes)
	if err != nil {
		return err
	}

	v1Svc.Handler.Match = httpCfg.AuthenticateURLsMatching
	v1Svc.Handler.Type = httpCfg.AuthenticationStrategy

	return nil
}
