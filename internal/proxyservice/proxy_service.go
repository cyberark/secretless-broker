package proxyservice

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

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
	for _, service := range s.config.Services {
		err := s.ensureSocketIsDeleted(service.ListenOn)
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			s.logger.Panic(err)
		}
	}

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

	s.logger.Infoln("Stopping services...")
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

	// TODO: bug
	// 1. We don't preemptively check if all plugins exist. This results in the implicit assumptios that
	// the plugin we can't find is a TCP one. Because of the way we categorise services

	// TCP Plugins
	for _, cfg := range s.configsByType.TCP {
		// Validation will have already happened
		tcpSvc, err := s.createTCPService(cfg, tcpPlugins[cfg.Connector])
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			s.logger.Panicf("unable to create TCP service '%s': %s", cfg.Name, err)
		}
		servicesToStart = append(servicesToStart, tcpSvc)
	}

	// HTTP Plugins
	for _, httpSvcConfig := range s.configsByType.HTTP {
		// Validation will have already happened
		httpSvc, err := s.createHTTPService(httpSvcConfig, httpPlugins)
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			s.logger.Panicf(
				"unable to create HTTP proxy service on '%s': %s",
				httpSvcConfig.SharedListenOn,
				err,
			)
		}
		servicesToStart = append(servicesToStart, httpSvc)
	}

	// SSH Plugins
	for _, cfg := range s.configsByType.SSH {
		// Validation will have already happened
		sshSvc, err := s.createSSHService(cfg)
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			s.logger.Panicf("unable to create SSH service '%s': %s", cfg.Name, err)
		}
		servicesToStart = append(servicesToStart, sshSvc)
	}

	// SSH Agent Plugins
	for _, cfg := range s.configsByType.SSHAgent {
		// Validation will have already happened
		sshAgentSvc, err := s.createSSHAgentService(cfg)
		if err != nil {
			// TODO: Add Fatalf to our logger and use that
			s.logger.Panicf("unable to create SSH Agent service '%s': %s", cfg.Name, err)
		}
		servicesToStart = append(servicesToStart, sshAgentSvc)
	}

	// TODO: Deal with SSH in a hardcoded way

	return servicesToStart
}

// If we are a socket listener and there is a socket file already
// we need to ensure that the socket file is remoed before starting
// the service.
func (s *proxyServices) ensureSocketIsDeleted(address string) error {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("unable to parse ListenOn location '%s'", address)
	}

	// If we're not a unix socket address, we don't need to worry about pre-emptive cleanup
	if parsedURL.Scheme != "unix" {
		return nil
	}

	socketFile := parsedURL.Path
	s.logger.Debugf("Ensuring that the socketfile '%s' is not present...", socketFile)

	// If file is not present, then we are ok to continue
	if _, err := os.Stat(socketFile); os.IsNotExist(err) {
		s.logger.Debugf("Socket file '%s' not present. Skipping deletion.", socketFile)
		return nil
	}

	// Otherwise delete the file first
	s.logger.Warnf("Socket file '%s' already present. Deleting...", socketFile)
	err = os.Remove(socketFile)
	if err != nil {
		return fmt.Errorf("unable to delete stale ocket file '%s'", socketFile)
	}

	return nil
}

// TODO: v2.HTTPServiceConfig is a value type.  It needs to be moved to a
//   separate package  All hardcoded deps that has no dependencies.
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

		// Get the http traffic patterns to match from the connector config.
		httpCfg, err := v2.NewHTTPConfig(subCfg.ConnectorConfig)
		if err != nil {
			return nil, err
		}

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
	netAddr := v2.NetworkAddress(config.ListenOn)
	listener, err := net.Listen(netAddr.Network(), netAddr.Address())
	if err != nil {
		return nil, err
	}

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
	netAddr := v2.NetworkAddress(config.ListenOn)
	listener, err := net.Listen(netAddr.Network(), netAddr.Address())
	if err != nil {
		return nil, err
	}

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
	netAddr := v2.NetworkAddress(config.ListenOn)
	listener, err := net.Listen(netAddr.Network(), netAddr.Address())
	if err != nil {
		return nil, err
	}

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
		return s.getSecrets(creds)
	}
}

// getSecrets returns the secret values for the requested credentials.
// TODO: Move this up one level, pass it down as dep.  Danger: This has
//   a hardcoded dependency on plugin and v1.
func (s *proxyServices) getSecrets(secrets []*v2.Credential) (map[string][]byte, error) {
	return s.resolver.Resolve(secrets)
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
