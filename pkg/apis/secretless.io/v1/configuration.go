package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
XXX: Ensure code generators are re-run anytime fields are added, removed, or
     their types changed!
*/

// Variable is a named secret.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Variable struct {
	meta_v1.TypeMeta   `json:",omitempty" yaml:",omitempty"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Name is the name by which the variable will be used by the client.
	Name string `json:"name,omitempty" yaml:",omitempty"`
	// Provider is the provider name.
	Provider string `json:"provider"`
	// Value is the identifier of the secret that the Provider will load.
	ID string `json:"id"`
}

// Listener listens on a port on socket for inbound connections, which are
// handed off to Handlers.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Listener struct {
	meta_v1.TypeMeta   `json:",omitempty" yaml:",omitempty"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Address     string   `json:"address,omitempty" yaml:",omitempty"`
	CACertFiles []string `yaml:"caCertFiles,omitempty" yaml:",omitempty"`
	Debug       bool     `json:"debug,omitempty" yaml:",omitempty"`
	Name        string   `json:"name"`
	Protocol    string   `json:"protocol"`
	Socket      string   `json:"socket,omitempty" yaml:",omitempty"`
}

// Handler processes an inbound message and connects to a specified backend
// using Credentials which it fetches from a provider.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Handler struct {
	meta_v1.TypeMeta   `json:",omitempty" yaml:",omitempty"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name         string     `json:"name,omitempty" yaml:",omitempty"`
	Type         string     `json:"type"`
	ListenerName string     `json:"listener" yaml:"listener"`
	Debug        bool       `json:"debug,omitempty" yaml:",omitempty"`
	Match        []string   `json:"match,omitempty" yaml:",omitempty"`
	Credentials  []Variable `json:"credentials,omitempty" yaml:",omitempty"`
}

// ConfigurationSpec is the main configuration structure for Secretless.
// It lists and configures the protocol listeners and handlers.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ConfigurationSpec struct {
	meta_v1.TypeMeta   `json:",omitempty" yaml:",omitempty"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Listeners []Listener `json:"listeners"`
	Handlers  []Handler  `json:"handlers"`
}

// Configuration is the generic CRD API type wrapping our spec
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Configuration struct {
	meta_v1.TypeMeta   `json:",omitempty" yaml:",omitempty"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   ConfigurationSpec   `json:"spec"`
	Status ConfigurationStatus `json:"status"`
}

// ConfigurationStatus is used to indicate what state the CRD is in
type ConfigurationStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// ConfigurationList is an array container of our CRD resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ConfigurationList struct {
	meta_v1.TypeMeta `json:",omitempty" yaml:",omitempty"`
	meta_v1.ListMeta `json:"metadata" yaml:"metadata,omitempty"`

	Items []Configuration `json:"items"`
}
