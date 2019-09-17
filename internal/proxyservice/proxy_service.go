package proxyservice

import (
	"github.com/cyberark/secretless-broker/internal/plugin"
	v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// TODO: move to impl package
type proxyServices struct {
	config        v2.Config
	logger        log.Logger
	eventNotifier v1.EventNotifier
	availPlugins  plugin.AvailablePlugins
}

// AvailPluginStub is a temporary placeholder for AvailablePlugins
type AvailPluginStub struct {}
// HTTPPlugins returns the available HTTP plugins.
func (ap *AvailPluginStub) HTTPPlugins() map[string]http.Plugin {
	return nil
}
// TCPPlugins returns the available TCP plugins.
func (ap *AvailPluginStub) TCPPlugins() map[string]tcp.Plugin {
	return nil
}

// TODO: Rename to Call or Run and return a Stopper instead of having Stop()
// Start starts all proxy services
func (s *proxyServices) Start() {
	// TODO: Implement

	// For each ProxyService:
	//
	// 1. Rewrap the Logger with service name prefix
	// 2. Create the ConnectorResources object
	// etc...
}

// Stop stops all proxy services
func (s *proxyServices) Stop() {
	// Stop all the Service
}

// NewProxyServices returns a new ProxyServices instance.
func NewProxyServices(
	cfg v2.Config,
	availPlugins plugin.AvailablePlugins,
	logger log.Logger,
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
