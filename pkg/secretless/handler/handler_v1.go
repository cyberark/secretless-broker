package handler

import (
	"net"

	"github.com/conjurinc/secretless/pkg/secretless/config"
)

// Handler is an interface which takes a connection and connects it to a backend
type Handler_v1 interface {
	GetConfig() config.Handler
	GetClientConnection() net.Conn
	GetBackendConnection() net.Conn
}
