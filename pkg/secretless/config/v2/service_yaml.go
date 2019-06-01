package v2

import (
	"sort"
)

type serviceYAML struct {
	Protocol    string          `yaml:"protocol" json:"protocol"`
	ListenOn    string          `yaml:"listenOn" json:"listenOn"`
	Credentials credentialsYAML `yaml:"credentials" json:"credentials"`
	Config      interface{}     `yaml:"config" json:"config"`
}

type servicesYAML map[string]*serviceYAML

func (servicesYAML *servicesYAML) ToServices() ([]*Service, error) {

	services := make([]*Service, 0)
	for svcName, svcYAML := range *servicesYAML {
		svc, err := NewService(svcName, svcYAML)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	// sort services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services, nil
}
