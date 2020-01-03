package mock

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// TCPPlugin is a mock struct that implements tcp.Plugin interface
type TCPPlugin struct {
	tcp.Plugin

	id string
}

// NewTCPPlugin creates a new TCPPlugin mock with an id, so that it may
// be distinguished from other mocks by DeepEqual.
func NewTCPPlugin(id string) TCPPlugin {
	return TCPPlugin{id: id}
}
