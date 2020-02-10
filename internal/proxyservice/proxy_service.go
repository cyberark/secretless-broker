package proxyservice

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/internal/plugin"
	httpproxy "github.com/cyberark/secretless-broker/internal/plugin/connectors/http"
	sshproxy "github.com/cyberark/secretless-broker/internal/plugin/connectors/ssh"
	sshagentproxy "github.com/cyberark/secretless-broker/internal/plugin/connectors/sshagent"
	tcpproxy "github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp"
	v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
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
	configsByType   v2.ConfigsByType
	eventNotifier   v1.EventNotifier
	logger          logapi.Logger
	resolver        v1.Resolver
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

	s.logger.Infoln("Stopping all services...")
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
	httpPlugins := s.availPlugins.HTTPPlugins()
	tcpPlugins := s.availPlugins.TCPPlugins()

	errors := validation.Errors{}

	// TCP Plugins
	for _, cfg := range s.configsByType.TCP {
		// Validation will have already happened
		tcpSvc, err := s.createTCPService(cfg, tcpPlugins[cfg.Connector])
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			errors[cfg.Name] = fmt.Errorf(
				"unable to create TCP service: %s",
				err,
			)
			continue
		}
		servicesToStart = append(servicesToStart, tcpSvc)
	}

	// HTTP Plugins
	for _, httpSvcConfig := range s.configsByType.HTTP {
		// Validation will have already happened
		httpSvc, err := s.createHTTPService(httpSvcConfig, httpPlugins)
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			errors[httpSvcConfig.Name()] = fmt.Errorf(
				"unable to create HTTP proxy service on '%s': %s",
				httpSvcConfig.SharedListenOn,
				err,
			)
			continue
		}
		servicesToStart = append(servicesToStart, httpSvc)
	}

	// SSH Plugins
	for _, cfg := range s.configsByType.SSH {
		// Validation will have already happened
		sshSvc, err := s.createSSHService(cfg)
		if err != nil {
			errors[cfg.Name] = fmt.Errorf(
				"unable to create SSH service: %s",
				err,
			)
			continue
		}
		servicesToStart = append(servicesToStart, sshSvc)
	}

	// SSH Agent Plugins
	for _, cfg := range s.configsByType.SSHAgent {
		// Validation will have already happened
		sshAgentSvc, err := s.createSSHAgentService(cfg)
		if err != nil {
			errors[cfg.Name] = fmt.Errorf(
				"unable to create SSH Agent service: %s",
				err,
			)
			continue
		}
		servicesToStart = append(servicesToStart, sshAgentSvc)
	}

	// If there are errors, we need to show them. This method exits the
	// program if any errors are detected.
	handleErrors(errors, s.logger)

	return servicesToStart
}

func handleErrors(errors validation.Errors, logger logapi.Logger) {
	if len(errors) > 0 {
		for cfgName, err := range errors {
			logger.Errorf("Fatal error in '%s': %s", cfgName, err)
		}

		os.Exit(1)
	}
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

	s.logger.Infof("Starting HTTP listener on %s...", netAddr.Address())

	// Create the subservices

	var subservices []httpproxy.Subservice
	for _, subCfg := range httpSvcCfg.SubserviceConfigs {
		// "cur" naming prefix needed to avoid package name collision
		curPlugin := plugins[subCfg.Connector]
		connResources := s.connectorResources(subCfg)
		curConnector := curPlugin.NewConnector(connResources)
		credsRetriever := s.credsRetriever(subCfg.Credentials)

		// Get the http traffic patterns to match from the connector config.
		httpCfg, err := v2.NewHTTPConfig(subCfg.ConnectorConfig)
		if err != nil {
			s.logger.Errorf("configuration parsing of '%s' failed: %s", subCfg.Connector, err)
			return nil, err
		}

		s.logger.Infof("Starting HTTP subservice %s...", subCfg.Connector)

		subservices = append(subservices, httpproxy.Subservice{
			ConnectorID:              subCfg.Connector, // TODO: Rename connectorID
			Connector:                curConnector,
			RetrieveCredentials:      credsRetriever,
			AuthenticateURLsMatching: httpCfg.AuthenticateURLsMatching,
		})
	}

	// Create the logger
	// HTTP proxy service gets its own logger (subservices have own loggers)

	proxyName := httpSvcCfg.Name()
	svcLogger := s.loggerFor(proxyName)

	// TODO: NewHTTPProxyFunc needs to be injected
	newSvc, err := httpproxy.NewProxyService(subservices, listener, svcLogger)
	if err != nil {
		s.logger.Errorf("could not create http proxy service '%s'", proxyName)
		return nil, err
	}
	return newSvc, nil
}

func (s *proxyServices) createSSHService(
	config v2.Service,
) (internal.Service, error) {

	// TODO: Add validation somewhere about overlapping listenOns
	// TODO: v2.NetworkAddress is a value type.  It needs to be moved to its
	//   own package with no deps (stdlib deps are ok).
	netAddr := config.ListenOn
	listener, err := net.Listen(netAddr.Network(), netAddr.Address())
	if err != nil {
		return nil, err
	}

	s.logger.Infof("Starting SSH listener on %s...", netAddr.Address())

	connResources := s.connectorResources(config)
	credsRetriever := s.credsRetriever(config.Credentials)

	// TODO: NewProxyService needs to be injected
	newSvc, err := sshproxy.NewProxyService(
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

func (s *proxyServices) createSSHAgentService(
	config v2.Service,
) (internal.Service, error) {

	// TODO: Add validation somewhere about overlapping listenOns
	// TODO: v2.NetworkAddress is a value type.  It needs to be moved to its
	//   own package with no deps (stdlib deps are ok).
	netAddr := config.ListenOn
	listener, err := net.Listen(netAddr.Network(), netAddr.Address())
	if err != nil {
		return nil, err
	}

	s.logger.Infof("Starting SSH Agent listener on %s...", netAddr.Address())

	connResources := s.connectorResources(config)
	credsRetriever := s.credsRetriever(config.Credentials)

	// TODO: NewProxyService needs to be injected
	newSvc, err := sshagentproxy.NewProxyService(
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

func (s *proxyServices) createTCPService(
	config v2.Service,
	pluginInst tcp.Plugin,
) (internal.Service, error) {

	// TODO: Add validation somewhere about overlapping listenOns
	// TODO: v2.NetworkAddress is a value type.  It needs to be moved to its
	//   own package with no deps (stdlib deps are ok).
	netAddr := config.ListenOn
	listener, err := net.Listen(netAddr.Network(), netAddr.Address())
	if err != nil {
		return nil, err
	}

	s.logger.Infof("Starting TCP listener on %s...", netAddr.Address())

	connResources := s.connectorResources(config)
	svcConnector := pluginInst.NewConnector(connResources)
	credsRetriever := s.credsRetriever(config.Credentials)

	// TODO: NewTCPProxyFunc needs to be injected
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
		return s.resolver.Resolve(creds)
	}
}

// NewProxyServices returns a new ProxyServices instance.
// TODO: Reconsider the Resolver design so it's exactly what we need for the new code.
// TODO: v1.Provider options should be an interface
func NewProxyServices(
	cfg v2.Config,
	availPlugins plugin2.AvailablePlugins,
	logger logapi.Logger,
	evtNotifier v1.EventNotifier,
) internal.Service {

	// Setup our resolver
	providerFactories := make(map[string]func(v1.ProviderOptions) (v1.Provider, error))

	for providerID, providerFactory := range providers.ProviderFactories {
		providerFactories[providerID] = providerFactory
	}

	resolver := plugin.NewResolver(providerFactories, nil, nil)

	// Create the proxyServices object
	services := proxyServices{
		availPlugins:  availPlugins,
		config:        cfg,
		eventNotifier: evtNotifier,
		logger:        logger,
		resolver:      resolver,
	}

	// TODO: v2.NewConfigsByType should be an interface, so we can remove this
	//   hardcoded dep on an impl type of v2.  All deps need to be injected.
	services.configsByType = v2.NewConfigsByType(cfg.Services, availPlugins)

	return &services
}
