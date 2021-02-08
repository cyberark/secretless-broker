package pg

import (
	"fmt"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/pg/protocol"
)

// Startup performs the startup handshake with the client and parses the client
// options to extract the database name.
func (s *SingleUseConnector) Startup() error {
	s.logger.Debugf(
		"Handling connection %+v -> %+v",
		s.clientConn.RemoteAddr(),
		s.clientConn.LocalAddr(),
	)

	version, options, err := s.parseStartupMessage()
	if err != nil {
		return err
	}

	// Handle the case where the startup message was an SSL request.
	if version == protocol.SSLRequestCode {
		s.sslNotSupported()
		version, options, err = s.parseStartupMessage()
		if err != nil {
			return err
		}

		// There should not be a second SSLRequest
		if version == protocol.SSLRequestCode {
			return fmt.Errorf("Unexpected SSL Request after SSL not supported response")
		}
	}

	var ok bool
	s.databaseName, ok = options["database"]
	if !ok {
		return fmt.Errorf("no 'database' found in connect options")
	}

	return nil
}

func (s *SingleUseConnector) parseStartupMessage() (int32, map[string]string, error) {
	messageBytes, err := protocol.ReadStartupMessage(s.clientConn)
	if err != nil {
		return 0, nil, err
	}

	version, options, err := protocol.ParseStartupMessage(messageBytes)
	if err != nil {
		return 0, nil, err
	}

	s.logger.Debugf(
		"s.Client version: %v, (SSL mode: %v)",
		version,
		version == protocol.SSLRequestCode)

	return version, options, nil
}
