package proxy_service

import (
	v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// TODO: move to impl package
type _secretless struct {
	config        v2.Config
	logger        log.Logger
	eventNotifier v1.EventNotifier
	availPlugins  secretless.AvailablePlugins
}

// TODO: This will be replaced by real impl with Srdjan's PR
type AvailPluginStub struct {}
func (ap *AvailPluginStub) HTTPPlugins() map[string]http.Plugin {
	return nil
}
func (ap *AvailPluginStub) TCPPlugins() map[string]tcp.Plugin {
	return nil
}

// TODO: Rename to Call or Run and return a Stopper instead of having Stop()
func (s *_secretless) Start() {
	// TODO: Implement

	// For each ProxyService:
	//
	// 1. Rewrap the Logger with service name prefix
	// 2. Create the ConnectorResources object
	// etc...
}

func (s *_secretless) Stop() {
	// Stop all the ProxyServices
}

// called in StartSecretless
func NewStartProxyServices(
	cfg v2.Config,
	availPlugins secretless.AvailablePlugins,
	logger log.Logger,
	evtNotifier v1.EventNotifier,
) secretless.StartProxyServices {

	secretlessObj := _secretless{
		config:        cfg,
		logger:        logger,
		eventNotifier: evtNotifier,
		availPlugins:  availPlugins,
	}

	// TODO: create our unstarted ProxyServices here
	//   logic uses availPlugins and config to figure out what services to start

	return &secretlessObj
}
