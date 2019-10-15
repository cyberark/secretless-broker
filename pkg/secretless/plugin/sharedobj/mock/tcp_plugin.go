package mock

import "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"

// TCPPlugin is a mock struct that implements tcp.Plugin interface
type TCPPlugin struct{ tcp.Plugin }
