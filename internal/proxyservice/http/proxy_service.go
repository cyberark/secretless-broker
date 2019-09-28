package http

import (
	"fmt"
	"net"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
)

// Subservice handles specific traffic within an HTTP proxy service, using
// traffic filtering rules and a devoted Connector.
type Subservice struct {
	Connector http.Connector
	RetrieveCredentials internal.CredentialsRetriever
}

// NewProxyService create a new HTTP proxy service.
// TODO: Replace this stub with real implementation:
//   https://github.com/cyberark/secretless-broker/issues/848
func NewProxyService(
	subservices []Subservice,
	sharedListener net.Listener,
	logger log.Logger,
) (internal.Service, error) {
	fmt.Println(subservices, sharedListener, logger)
	return nil,nil
}
