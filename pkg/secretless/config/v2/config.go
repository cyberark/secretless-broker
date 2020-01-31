package v2

import (
	"fmt"
	"log"
	"sort"

	"gopkg.in/yaml.v2"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/sharedobj"
)

// Config represents a full configuration of Secretless, which is just a list of
// individual Service configurations.
type Config struct {
	Debug    bool
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

// MarshalYAML serializes Config to the secretless.yml format
func (c Config) MarshalYAML() (interface{}, error) {
	servicesAsYAML := map[string]*serviceYAML{}
	for _, svc := range c.Services {
		credentialYamls := credentialsYAML{}
		for _, cred := range svc.Credentials {
			credentialYamls[cred.Name] = struct {
				From string `yaml:"from" json:"from"`
				Get  string `yaml:"get" json:"get"`
			}{
				From: cred.From,
				Get:  cred.Get,
			}
		}

		servicesAsYAML[svc.Name] = &serviceYAML{
			Connector:   svc.Connector,
			ListenOn:    string(svc.ListenOn),
			Credentials: credentialYamls,
			Config:      svc.ConnectorConfig,
		}
	}

	return struct {
		Version  string                  `yaml:"version" json:"version"`
		Services map[string]*serviceYAML `yaml:"services" json:"services"`
	}{
		Version:  "2",
		Services: servicesAsYAML,
	}, nil
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

	var httpConfigs, tcpConfigs, sshConfigs, sshAgentConfigs []Service

	for _, cfg := range rawConfigs {
		switch {
		case sharedobj.IsHTTPPlugin(availPlugins, cfg.Connector):
			httpConfigs = append(httpConfigs, cfg)
			continue
		case cfg.Connector == "ssh":
			sshConfigs = append(sshConfigs, cfg)
			continue
		case cfg.Connector == "ssh-agent":
			sshAgentConfigs = append(sshAgentConfigs, cfg)
			continue
		default:
			tcpConfigs = append(tcpConfigs, cfg)
		}
	}

	httpByListenOn := groupedByListenOn(httpConfigs)

	// Now create proper HTTPServiceConfig objects from our map
	var httpServiceConfigs []HTTPServiceConfig
	for listenOn, configs := range httpByListenOn {
		httpServiceConfig := HTTPServiceConfig{
			SharedListenOn:    listenOn,
			SubserviceConfigs: configs,
		}
		httpServiceConfigs = append(httpServiceConfigs, httpServiceConfig)
	}

	return ConfigsByType{
		HTTP:     httpServiceConfigs,
		SSH:      sshConfigs,
		SSHAgent: sshAgentConfigs,
		TCP:      tcpConfigs,
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
	HTTP     []HTTPServiceConfig
	SSH      []Service
	SSHAgent []Service
	TCP      []Service
}

// groupedByListenOn returns a map grouping the configs provided by their ListenOn
// property.  Merely a helper function to reduce bloat in newConfigsByType.
func groupedByListenOn(httpConfigs []Service) map[NetworkAddress][]Service {
	httpByListenOn := map[NetworkAddress][]Service{}
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
