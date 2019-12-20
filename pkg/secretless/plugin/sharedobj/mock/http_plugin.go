package mock

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
)

// HTTPPlugin is a mock struct that implements http.Plugin interface
type HTTPPlugin struct {
	http.Plugin

	id string
}

// NewHTTPPlugin creates a new HTTPPlugin mock with an id, so that it may
// be distinguished from other mocks by DeepEqual.
func NewHTTPPlugin(id string) HTTPPlugin {
	return HTTPPlugin{id: id}
}
