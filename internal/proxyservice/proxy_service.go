package proxyservice

import (
	"fmt"
	"net"
	"strings"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/internal/plugin"
	v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	tcpproxy "github.com/cyberark/secretless-broker/internal/proxyservice/tcp"
	"github.com/cyberark/secretless-broker/pkg/secretless"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	logapi "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// TODO: move to impl package
type proxyServices struct {
	config          v2.Config
	logger          logapi.Logger
	eventNotifier   v1.EventNotifier
	availPlugins    plugin.AvailablePlugins
	runningServices []secretless.Service
}

// Start starts all proxy services
func (s *proxyServices) Start() error {
	for _, svc := range s.servicesToStart() {
		err := svc.Start()
		// Showstopper: Failure to start a service is fatal.
		if err != nil {
			// TODO: Upgrade our logger so we can use Fatalf here
			s.logger.Panicf("could not start proxy service: %s", err)
		}
		s.runningServices = append(s.runningServices, svc)
	}
	return nil
}

// Stop stops all proxy services
func (s *proxyServices) Stop() error {
	var stopFailures []string
	for _, svc := range s.runningServices {
		err := svc.Stop()
		if err != nil {
			s.logger.Errorf("could not stop proxy service: %s", err)
			stopFailures = append(stopFailures, err.Error())
		}
	}

	if len(stopFailures) > 0 {
		return fmt.Errorf(
			"these errors occured while stopping all services: %s",
			strings.Join(stopFailures, "; "),
		)
	}
	return nil
}

func (s *proxyServices) servicesToStart() []secretless.Service {
	var servicesToStart []secretless.Service

	tcpPlugins := s.availPlugins.TCPPlugins()
	// httpPlugins := s.availPlugins.HTTPPlugins()

	for _, svc := range s.config.Services {
		requestedPlugin := svc.Connector //TODO: this rename is a name smell

		// first check the available TCP Plugins
		tcpPlugin, found := tcpPlugins[requestedPlugin]
		if found {
			tcpSvc, err := s.createTCPService(svc, tcpPlugin)
			// Failure to create is fatal.
			if err != nil {
				// TODO: Add Fatalf to our logger and use that
				s.logger.Panicf("unable to create TCP service '%s'", svc.Name)
			}
			servicesToStart = append(servicesToStart, tcpSvc)
		}

		// TODO: next check available HTTP Plugins
		// httpPlugin, found := httpPlugins[requestedPlugin]

		// TODO: Deal with SSH in a hardcoded way

		// Default case: not found
		// TODO: Add Fatalf to our logger and use that
		s.logger.Panicf("plugin '%s' not available.", requestedPlugin)
	}
	return servicesToStart
}

func (s *proxyServices) createTCPService(
	svc *v2.Service,
	plugin tcp.Plugin,
) (secretless.Service, error) {

	//TODO: Add validation somewhere about overlapping listenOns
	listener, err := net.Listen("tcp", strings.TrimLeft(svc.ListenOn, "tcp://"))
	if err != nil {
		s.logger.Errorf("could not create listener on: %s", svc.ListenOn)
		return nil, err
	}

	svcLogger := s.logger.CopyWith(svc.Name, s.logger.DebugEnabled())
	connResources := connector.NewResources(svc.ConnectorConfig, svcLogger)
	svcConnector := plugin.NewConnector(connResources)

	// Temp var required so that the function closes over the current
	// loop value.
	credsCopy := svc.Credentials
	credsRetriever := func() (map[string][]byte, error) {
		return GetSecrets(credsCopy)
	}

	newSvc, err := tcpproxy.NewProxyService(
		svcConnector,
		listener,
		svcLogger,
		credsRetriever,
	)

	if err != nil {
		s.logger.Errorf("could not create proxy service '%s'", svc.Name)
		return nil, err
	}

	return newSvc, nil
}

// NewProxyServices returns a new ProxyServices instance.
func NewProxyServices(
	cfg v2.Config,
	availPlugins plugin.AvailablePlugins,
	logger logapi.Logger,
	evtNotifier v1.EventNotifier,
) secretless.Service {

	secretlessObj := proxyServices{
		config:        cfg,
		logger:        logger,
		eventNotifier: evtNotifier,
		availPlugins:  availPlugins,
	}

	// TODO: create our unstarted Service here
	//   logic uses availPlugins and config to figure out what services to start

	return &secretlessObj
}

// GetSecrets returns the secret values for the requested credentials.
// TODO: Move this up one level, pass it down as dep.  Also, reconsider the
//   Resolver design so it's exactly what we need for the new code.
func GetSecrets(secrets []*v2.Credential) (map[string][]byte, error) {
	providerFactories := make(map[string]func(v1.ProviderOptions) (v1.Provider, error))

	for providerID, providerFactory := range internal.InternalProviders {
		providerFactories[providerID] = providerFactory
	}

	resolver := plugin.NewResolver(providerFactories, nil, nil)

	return resolver.Resolve(secrets)
}
