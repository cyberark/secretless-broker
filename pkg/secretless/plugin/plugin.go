package plugin

import (
	"github.com/conjurinc/secretless/pkg/secretless/handler"
	"github.com/conjurinc/secretless/pkg/secretless/listener"
	"github.com/conjurinc/secretless/pkg/secretless/manager"
)

// ----------------- V1 interfaces -----------------
// Exported v1 interfaces
type Listener_v1 = listener.Listener_v1
type Handler_v1 = handler.Handler_v1
type Manager_v1 = manager.Manager_v1
