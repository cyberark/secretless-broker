/*
Package proxyservice takes a Secretless configuration and available plugins
and constructs the requires ProxyServices that Secretless will run.

To put this in context, the complete high-level flow is:

	1. Parse the `secretless.yml` into individual “service” configs,
	corresponding to each service entry in the yml.

	2. Identify http services that share a `listenOn`, because we know those
	will all be part of a single “http proxy service” that uses traffic routing
	within it to delegate to the subservice connectors.

	3. So now we have have the http service's `listenOn` and all the
	“subservices” associated with it.

	4. Each of those subservices needs two things: a connector (which knows how
	authenticate requests) and a way to get the current credentials at runtime.

	5. Now note the signature of the connector itself just looks like this:

		type Connector func(request *http.Request, secrets plugin.SecretsByID) error

	6. So it is the responsibility of the proxy service to actually fetch the
	credentials.  So what does the proxy service need for each of those
	subservices? Precisely this:

		type HTTPSubService struct {
		  connector http.Connector,
		  retrieveCredentials internal.CredentialsRetriever,
		}

	7. Putting all the together, here’s what we need to construct a new http
	proxy service:

		func NewProxyService(
			subservices []HTTPSubService,
			sharedListener net.Listener,
			logger log.Logger,
		) (internal.Service, error) {

One fine point that’s not be obvious: Each of the subservice connectors gets
created with its own custom logger, but those A. those aren’t accessible to the
proxy service itself and B. even if they were, we’d want to explicitly pass the
logger to be clear about this is a dependency of the proxy service itself and C.
the proxy service’s logger should have a different prefix than those of the more
specific subservices.  So that’s why we’re passing the logger too.

TODO: Add a principled explanation about how logging is working.
  Eg, when do we return errors, when is it fatal?  what do we log, everything?
TODO: This is long, perhaps move it into doc.go
 */
package proxyservice

import (
	"fmt"
	"net"
	"strings"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/internal/plugin"
	v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	httpproxy "github.com/cyberark/secretless-broker/internal/proxyservice/http"
	tcpproxy "github.com/cyberark/secretless-broker/internal/proxyservice/tcp"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	logapi "github.com/cyberark/secretless-broker/pkg/secretless/log"
	plugin2 "github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// TODO: move to impl package
type proxyServices struct {
	availPlugins    plugin2.AvailablePlugins
	config          v2.Config
	eventNotifier   v1.EventNotifier
	logger          logapi.Logger
	runningServices []internal.Service
}

// Start starts all proxy services
func (s *proxyServices) Start() error {
	for _, svc := range s.servicesToStart() {
		err := svc.Start()
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

func (s *proxyServices) servicesToStart() (servicesToStart []internal.Service) {
	configsByType := v2.NewConfigsByType(s.config.Services, s.availPlugins)
	httpPlugins := s.availPlugins.HTTPPlugins()
	tcpPlugins := s.availPlugins.TCPPlugins()

	// TCP Plugins
	for _, cfg := range configsByType.TCP {
		// Validation will have already happened
		tcpSvc, err := s.createTCPService(cfg, tcpPlugins[cfg.Connector])
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			s.logger.Panicf("unable to create TCP service '%s'", cfg.Name)
		}
		servicesToStart = append(servicesToStart, tcpSvc)
	}

	// HTTP Plugins
	for _, httpSvcConfig := range configsByType.HTTP {
		// Validation will have already happened
		httpSvc, err := s.createHTTPService(httpSvcConfig, httpPlugins)
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			s.logger.Panicf(
				"unable to create HTTP proxy service on: '%s'",
				httpSvcConfig.SharedListenOn,
			)
		}
		servicesToStart = append(servicesToStart, httpSvc)
	}

	// TODO: Deal with SSH in a hardcoded way

	return servicesToStart
}

func (s *proxyServices) createHTTPService(
	httpSvcCfg v2.HTTPServiceConfig,
	plugins map[string]http.Plugin,
) (internal.Service, error) {

	// Create the listener
	// TODO: If we want to unit test this, we'll need to inject net.Listen

	netAddr := httpSvcCfg.SharedListenOn
	listener, err := net.Listen(netAddr.Network(), netAddr.Address())
	if err != nil {
		s.logger.Errorf("listener creation failed: %s", httpSvcCfg.SharedListenOn)
		return nil, err
	}

	// Create the subservices

	var subservices []httpproxy.Subservice
	for _, subCfg := range httpSvcCfg.SubserviceConfigs {
		// "cur" naming prefix needed to avoid package name collision
		curPlugin := plugins[subCfg.Connector]
		connResources := s.connectorResources(subCfg)
		curConnector := curPlugin.NewConnector(connResources)
		credsRetriever := s.credsRetriever(subCfg.Credentials)

		subservices = append(subservices, httpproxy.Subservice{
			Connector:           curConnector,
			RetrieveCredentials: credsRetriever,
		})
	}

	// Create the logger
	// HTTP proxy service gets its own logger (subservices have own loggers)

	proxyName := httpSvcCfg.Name()
	svcLogger := s.loggerFor(proxyName)

	newSvc, err := httpproxy.NewProxyService(subservices, listener, svcLogger)
	if err != nil {
		s.logger.Errorf("could not create http proxy service '%s'", proxyName)
		return nil, err
	}
	return newSvc, nil
}

func (s *proxyServices) createTCPService(
	config v2.Service,
	plugin tcp.Plugin,
) (internal.Service, error) {

	//TODO: Add validation somewhere about overlapping listenOns
	netAddr := v2.NetworkAddress(config.ListenOn)
	listener, err := net.Listen(netAddr.Network(), netAddr.Address())
	if err != nil {
		s.logger.Errorf("could not create listener on: %s", config.ListenOn)
		return nil, err
	}

	connResources := s.connectorResources(config)
	svcConnector := plugin.NewConnector(connResources)
	credsRetriever := s.credsRetriever(config.Credentials)

	newSvc, err := tcpproxy.NewProxyService(
		svcConnector,
		listener,
		connResources.Logger(),
		credsRetriever,
	)

	if err != nil {
		s.logger.Errorf("could not create proxy service '%s'", config.Name)
		return nil, err
	}

	return newSvc, nil
}

func (s *proxyServices) connectorResources(svc v2.Service) connector.Resources {
	svcLogger := s.loggerFor(svc.Name)
	return connector.NewResources(svc.ConnectorConfig, svcLogger)
}

func (s *proxyServices) loggerFor(name string) logapi.Logger {
	return s.logger.CopyWith(name, s.logger.DebugEnabled())
}

func (s *proxyServices) credsRetriever(
	creds []*v2.Credential,
) internal.CredentialsRetriever {
	return func() (map[string][]byte, error) {
		return GetSecrets(creds)
	}
}

// NewProxyServices returns a new ProxyServices instance.
func NewProxyServices(
	cfg v2.Config,
	availPlugins plugin2.AvailablePlugins,
	logger logapi.Logger,
	evtNotifier v1.EventNotifier,
) internal.Service {

	secretlessObj := proxyServices{
		config:        cfg,
		logger:        logger,
		eventNotifier: evtNotifier,
		availPlugins:  availPlugins,
	}

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
