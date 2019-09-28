package v2

import (
	"fmt"
	"log"
	"sort"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/so"
	"gopkg.in/yaml.v2"
)

// Config represents a full configuration of Secretless, which is just a list of
// individual Service configurations.
type Config struct {
	Debug bool
	Services []*Service
}

// Serialize Config to YAML
func (c Config) String() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewConfig creates a v2.Config from yaml bytes
func NewConfig(v2YAML []byte) (*Config, error) {
	cfgYAML, err := newConfigYAML(v2YAML)
	if err != nil {
		return nil, err
	}

	services := make([]*Service, 0)
	for svcName, svcYAML := range cfgYAML.Services {
		svc, err := NewService(svcName, svcYAML)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}

	// sort Services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return &Config{
		Services: services,
	}, nil
}

// NewConfigsByType converts a slice of v2.Service configs into the configs
// needed to actually created ProxyServices -- configsByType.  In particular, it
// takes all the http configs and creates proper HTTPServiceConfig objects out
// of them -- grouping the raw v2.Service configs by their listenOn property.
// The remaining services are tcp, and already correspond 1-1 to the services
// we'll run.
// TODO: Eventually the application code should not be dealing directly with
//   []Service at all, but the processing into these more appropriate domain
//   configs should occur entirely at the border.
func NewConfigsByType(
	uncheckedConfigs []*Service,
	availPlugins plugin.AvailablePlugins,
) ConfigsByType {

	// Get the nil checks out of the way.
	var rawConfigs []Service
	for _, cfg := range uncheckedConfigs {
		if cfg == nil {
			// Hard-coding log here is okay since nils should never occur and
			// we won't be unit testing this.  Instead, we'll make the change
			// in the TODO_ above and this code will be deleted then.
			log.Fatalln("Nil configuration is not allowed!")
		}
		rawConfigs = append(rawConfigs, *cfg)
	}

	httpConfigs, tcpConfigs := separatedHTTPAndTCPConfigs(rawConfigs, availPlugins)
	httpByListenOn := groupedByListenOn(httpConfigs)

	// Now create proper HTTPServiceConfig objects from our map
	var httpServiceConfigs []HTTPServiceConfig
	for listenOn, configs := range httpByListenOn {
		httpServiceConfig := HTTPServiceConfig{
			SharedListenOn:    NetworkAddress(listenOn),
			SubserviceConfigs: configs,
		}
		httpServiceConfigs = append(httpServiceConfigs, httpServiceConfig)
	}

	return ConfigsByType{
		TCP:  tcpConfigs,
		HTTP: httpServiceConfigs,
	}
}

// HTTPServiceConfig represents an HTTP proxy service configuration. Multiple
// http entries within a v2.Service config slice that share a listenOn actually
// represent a single HTTP proxy service, with sub-handlers for different
// traffic.  This type captures that fact.
type HTTPServiceConfig struct {
	SharedListenOn    NetworkAddress
	SubserviceConfigs []Service
}

// Name returns the name of an HTTPServiceConfig
func (cfg *HTTPServiceConfig) Name() string {
	return fmt.Sprintf("HTTP Proxy on %s", cfg.SharedListenOn)
}

// ConfigsByType holds proxy service configuration in a form that directly
// corresponds to the ProxyService objects we want to create.  One ProxyService
// will be created for each entry in http, and one for each entry in tcp.
type ConfigsByType struct {
	TCP  []Service
	HTTP []HTTPServiceConfig
}

// separatedHTTPAndTCPConfigs takes a slices of configs and returns two slices,
// one containing only HTTP configs and the other containing only TCP configs.
// Merely a helper function to reduce bloat in newConfigsByType.
// TODO: There _might_ be something a little funny about the dependency here
//   on AvailablePlugins, but I'm not sure.  There are also reason for it.
//   Should consider this more fully.
func separatedHTTPAndTCPConfigs(
	configs []Service,
	availPlugins plugin.AvailablePlugins,
) (httpConfigs []Service, tcpConfigs []Service) {
	// TODO: Add proper validation here of the type.  This should moved into
	//   IsHTTPPlugin, whose API will likely change to returning a type or an
	//   error
	for _, cfg := range configs {
		if so.IsHTTPPlugin(availPlugins, cfg.Connector) {
			httpConfigs = append(httpConfigs, cfg)
			continue
		}
		tcpConfigs = append(tcpConfigs, cfg)
	}
	return httpConfigs, tcpConfigs
}

// groupedByListenOn returns a map grouping the configs provided by their ListenOn
// property.  Merely a helper function to reduce bloat in newConfigsByType.
func groupedByListenOn(httpConfigs []Service) map[string][]Service {
	httpByListenOn := map[string][]Service{}
	for _, httpConfig := range httpConfigs {
		// default group for this ListenOn, in case we don't yet have one yet
		var groupedConfigs []Service
		// but replace it with the existing group, if one exists
		for listenOn, alreadyGrouped := range httpByListenOn {
			if listenOn == httpConfig.ListenOn {
				groupedConfigs = alreadyGrouped
				break
			}
		}
		// append the current config to this ListenOn group
		groupedConfigs = append(groupedConfigs, httpConfig)
		httpByListenOn[httpConfig.ListenOn] = groupedConfigs
	}
	return httpByListenOn
}

