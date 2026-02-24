package troubleshooting

import (
	"io"
	"net"

	"github.com/cyberark/secretless-broker/internal/log"
	httpInternal "github.com/cyberark/secretless-broker/internal/plugin/connectors/http"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/http/generic"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// proxyService is a Secretless proxy service endpoint. The start logic of this service
// endpoint is abstracted so that the user can decide how it is implemented, common
// examples are in-process as a TCP proxy service or out-of-process as the Secretless
// binary.
type proxyService struct {
	host  string
	port  string
	start func() // start runs the logic to start the proxy service
	stop  func() // stop runs the logic to stop the proxy service
}

// Start concurrently runs the start logic for a proxy service
func (s *proxyService) Start() {
	s.start()
}

// Stop delegates the cleanup logic for a proxy service
func (s *proxyService) Stop() {
	s.stop()
}


// cloneCredentials creates an independent clone of a credentials map. The resulting
// clone will not be affected by any mutations to the original, and vice-versa. The clone
// is useful for passing to a proxyService service, to avoid zeroization of the original.
func cloneCredentials(original map[string][]byte) map[string][]byte {
	credsClone := make(map[string][]byte)

	for key, value := range original {
		// Clone the value
		valueClone := make([]byte, len(value))
		copy(valueClone, value)

		// Set the key, value pair on the credentials clone
		credsClone[key] = valueClone
	}

	return credsClone
}

// newInProcessProxyService creates an HTTP proxy service.
// 1. Create the net.Listener
// 2. Create the HTTP Service Connector
// 3. Create a TCP proxy service that uses (1) and (2)
func newInProcessProxyService(
	config []byte,
	credentials map[string][]byte,
	writer io.Writer,
) (*proxyService, error) {
	logger := log.NewWithOptions(writer, "", true)

	// Create a net.Listener on a random port
	listener, err := ephemeralListenerOnPort("0")
	if err != nil {
		return nil, err
	}

	// Extract the host and port from the net.Listener
	host, port, err := net.SplitHostPort(
		listener.Addr().String(),
	)
	if err != nil {
		return nil, err
	}

	// Create HTTP service connector
	svcConnector := generic.NewConnector(
		connector.NewResources(config, logger),
	)

	httpCfg, err := v2.NewHTTPConfig(config)
	if err != nil {
		return nil, err
	}

	// Create the TCP proxy service
	tcpProxySvc, err := httpInternal.NewProxyService(
		[]httpInternal.Subservice{
			{
				Connector:                svcConnector,
				ConnectorID:              "test",
				RetrieveCredentials:      func() (bytes map[string][]byte, e error) {
					// Clone credentials to prevents any mutation or zeroization
					return cloneCredentials(credentials), nil
				},
				AuthenticateURLsMatching: httpCfg.AuthenticateURLsMatching,
			},
		},
		listener,
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &proxyService{
		host: host,
		port: port,
		// Starts the TCP proxy service
		start: func() {
			err := tcpProxySvc.Start()
			if err != nil {
				logger.Warnf("proxyService#start: %s", err)
			}
		},
		// Stops the TCP proxy service, and cleans up
		stop: func() {
			// Ensure the proxyService service is stopped
			err = tcpProxySvc.Stop()
			if err != nil {
				logger.Warnf("proxyService#stop: %s", err)
			}
		},
	}, nil
}
