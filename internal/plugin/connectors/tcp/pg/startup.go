package pg

import (
	"fmt"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/pg/protocol"
)

// Startup performs the startup handshake with the client and parses the client
// options to extract the database name.
func (s *SingleUseConnector) Startup() error {
	s.logger.Debugln("Handling connection %v", s.clientConn)

	messageBytes, err := protocol.ReadStartupMessage(s.clientConn)
	if err != nil {
		return err
	}

	version, options, err := protocol.ParseStartupMessage(messageBytes);
	if  err != nil {
		return err
	}

	s.logger.Debugln(
		"s.Client version : %v, (SSL mode: %v)",
		version,
		version == protocol.SSLRequestCode)

	// Handle the case where the startup message was an SSL request.
	if version == protocol.SSLRequestCode {
		return fmt.Errorf("SSL not supported")
	}

	var ok bool
	s.databaseName, ok = options["database"]
	if !ok {
		return fmt.Errorf("no 'database' found in connect options")
	}

	return nil
}
