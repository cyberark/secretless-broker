package mssqltest

import (
	"net"

	"github.com/cyberark/secretless-broker/internal/log"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql"
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

// newInProcessProxyService creates an MSSQL TCP proxy service.
// 1. Create the net.Listener
// 2. Create the MSSQL Service Connector
// 3. Create a TCP proxy service that uses (1) and (2)
func newInProcessProxyService(
	credentials map[string][]byte,
) (*proxyService, error) {
	logger := log.New(true)

	// Create a net.Listener on a random port
	listener, err := localListenerOnPort("0")
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

	// Create MSSQL service connector
	svcConnector := mssql.NewConnector(
		connector.NewResources(nil, logger),
	)

	// Create the TCP proxy service
	tcpProxySvc, err := tcp.NewProxyService(
		svcConnector,
		listener,
		logger,
		func() (bytes map[string][]byte, e error) {
			// Clone credentials to prevents any mutation or zeroization
			return cloneCredentials(credentials), nil
		},
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
