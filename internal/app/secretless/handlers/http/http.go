package http

import (
	"log"
	"net/http"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// HttpSubHandler provides a a common interface for handlers that operate on Http connections
type HttpSubHandler interface {
	Authenticate(map[string][]byte, *http.Request) error
}

// HttpSubHandler applies authentication to the HTTP Authorization header.
type HttpHandler struct {
	plugin_v1.BaseHandler
	SubHandler HttpSubHandler
}

// HttpHandler instantiates a handler given HandlerOptions
func HttpHandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &HttpHandler{
		BaseHandler: plugin_v1.NewBaseHandler(options),
	}

	return handler
}

func (h *HttpHandler) RegisterSubHandler(subhandlerName string) {
	newSubHandler, ok := SubHandlers[subhandlerName]
	// Ensure that we have this sub-handler
	if !ok {
		log.Panicf("Error! Unrecognized handler id 'http/%s'", subhandlerName)
	}

	h.SubHandler = newSubHandler()

	if subhandlerName == "conjur" {
		// Force instantiate the Conjur provider so we can use an access token.
		// This will fail unless a means of authentication to Conjur is available.
		if h.Resolver != nil {
			h.Resolver.Provider("conjur");
		}
	}
}
