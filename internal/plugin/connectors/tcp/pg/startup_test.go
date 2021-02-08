package pg

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/pg/protocol"
	loggermock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
)

func readMessageType(r io.Reader) (byte, error) {
	messageTypeBytes := make([]byte, 1)
	if err := binary.Read(r, binary.BigEndian, &messageTypeBytes); err != nil {
		return 0, err
	}
	return messageTypeBytes[0], nil
}

func sendStartupMessage(w io.Writer, version int32) {
	message := protocol.CreateStartupMessage(
		version,
		"username-from-client",
		"db-from-client",
		map[string]string{
			"option-1": "option-1-from-client",
		})
	_, err := w.Write(message)
	if err != nil {
		panic(err)
	}
}

func pipeWithDeadline() (net.Conn, net.Conn) {
	r, w := net.Pipe()
	r.SetDeadline(time.Now().Add(2 * time.Second))
	w.SetDeadline(time.Now().Add(2 * time.Second))

	return r, w
}

func TestStartup(t *testing.T) {
	t.Run("handle client startup message with no TLS request", func(t *testing.T) {
		r, w := pipeWithDeadline()

		connector := SingleUseConnector{
			clientConn: r,
			logger:     &loggermock.LoggerMock{},
		}
		go func() {
			sendStartupMessage(w, protocol.ProtocolVersion)
		}()

		err := connector.Startup()
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, err, nil)
		assert.Equal(t, connector.databaseName, "db-from-client")
	})

	t.Run("handle client startup message with TLS request yielding to no TLS", func(t *testing.T) {
		r, w := pipeWithDeadline()

		connector := SingleUseConnector{
			clientConn: r,
			logger:     &loggermock.LoggerMock{},
		}
		go func() {
			sendStartupMessage(w, protocol.SSLRequestCode)

			messageType, err := readMessageType(w)
			assert.NoError(t, err)
			assert.Equal(t, messageType, protocol.SSLNotAllowed)

			sendStartupMessage(w, protocol.ProtocolVersion)
		}()

		err := connector.Startup()
		assert.NoError(t, err)
	})

	t.Run("error on client startup message with sustained TLS request after no TLS", func(t *testing.T) {
		r, w := pipeWithDeadline()

		connector := SingleUseConnector{
			clientConn: r,
			logger:     &loggermock.LoggerMock{},
		}
		go func() {
			sendStartupMessage(w, protocol.SSLRequestCode)

			messageType, err := readMessageType(w)
			assert.Nil(t, err)
			assert.Equal(t, messageType, protocol.SSLNotAllowed)

			sendStartupMessage(w, protocol.SSLRequestCode)

			messageType, err = readMessageType(w)
			assert.Nil(t, err)
			assert.Equal(t, messageType, protocol.SSLNotAllowed)
		}()

		err := connector.Startup()
		if !assert.Error(t, err) {
			return
		}
		assert.Contains(t, err.Error(), "Unexpected SSL Request after SSL not supported response")
	})
}
