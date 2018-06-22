package pg

import (
	"fmt"
	"log"

	"github.com/conjurinc/secretless/internal/app/secretless/listeners/pg/protocol"
)

// newClientOptions builds a ClientOptions from an options map.
// An error is returned if any required options are missing.
func newClientOptions(options map[string]string) (co *ClientOptions, err error) {
	co = &ClientOptions{Options: options}

	var ok bool
	co.User, ok = co.Options["user"]
	if !ok {
		err = fmt.Errorf("No 'user' found in connect options")
		return
	}
	co.Database, ok = co.Options["database"]
	if !ok {
		err = fmt.Errorf("No 'database' found in connect options")
		return
	}
	return
}

// Startup performs the startup handshake with the client and parses the ClientOptions.
func (h *Handler) Startup() (err error) {
	if h.Config.Debug {
		log.Printf("Handling connection %v", h.Client)
	}

	var messageBytes []byte
	if messageBytes, err = protocol.ReadStartupMessage(h.Client); err != nil {
		return
	}

	var version int32
	var options map[string]string
	if version, options, err = protocol.ParseStartupMessage(messageBytes); err != nil {
		return
	}

	if h.Config.Debug {
		log.Printf("h.Client version : %v, (SSL mode: %v)", version, version == protocol.SSLRequestCode)
	}

	// Handle the case where the startup message was an SSL request.
	if version == protocol.SSLRequestCode {
		err = fmt.Errorf("SSL not supported")
		return
	}

	h.ClientOptions, err = newClientOptions(options)
	if err != nil {
		return
	}

	return
}
