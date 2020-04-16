package tcp

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/mock"
	loggerMock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
)

const testString1 = "good heavens"
const testString2 = "moderate heavens"
const testString3 = "best of heavens"

func TestNewProxyService(t *testing.T) {
	t.Run("empty constructor arguments result in errors", func(t *testing.T) {
		_, err := NewProxyService(
			nil,
			nil,
			nil,
			nil,
		)
		assert.Error(t, err)
		if err == nil {
			return
		}
	})

	t.Run("non-empty constructor arguments result in no error", func(t *testing.T) {
		_, err := NewProxyService(
			mock.NewConnector(),
			mock.NewListener(),
			loggerMock.NewLogger(),
			mock.NewCredentialRetriever().RetrieveCredentials,
		)
		assert.NoError(t, err)
		if err != nil {
			return
		}
	})

	t.Run("zeroizes credentials from retrieveCredentials", func(t *testing.T) {
		// prepare
		connector := mock.NewConnector()

		credentialRetriever := mock.NewCredentialRetriever()
		creds := map[string][]byte{
			"credName": []byte("credValue"),
		}
		credentialRetriever.On("RetrieveCredentials").Return(
			creds,
			fmt.Errorf("couldn't retrieve credentials"))

		listener := mock.NewListener()

		// exercise
		ps := proxyService{
			connector:           connector,
			retrieveCredentials: credentialRetriever.RetrieveCredentials,
			listener:            listener,
		}
		err := ps.handleConnection(nil)

		// assert
		assert.Error(t, err)
		if err == nil {
			return
		}
		assert.Equal(
			t,
			strings.Repeat("\x00", len([]byte("credValue"))),
			string(creds["credName"]),
		)
	})
}

func TestProxyService_Start(t *testing.T) {
	t.Run("stopped proxy service cannot be restarted", func(t *testing.T) {
		connector := mock.NewConnector()
		credentialRetriever := mock.NewCredentialRetriever()
		listener := mock.NewListener()
		listener.On("Close").Return(nil)
		logger := loggerMock.NewLogger()

		ps, _ := NewProxyService(
			connector,
			listener,
			logger,
			credentialRetriever.RetrieveCredentials,
		)

		err := ps.Stop()
		assert.NoError(t, err)
		if err != nil {
			return
		}

		err = ps.Start()
		assert.Error(t, err)
	})

	t.Run("propagates error from Accept", func(t *testing.T) {
		// prepare
		logger := loggerMock.NewLogger()
		logger.On("Errorf").Return()

		connector := mock.NewConnector()
		connector.On("Connect").Return(nil, nil)

		credentialRetriever := mock.NewCredentialRetriever()
		credentialRetriever.On(
			"RetrieveCredentials",
		).Return(nil, nil)

		listener := mock.NewListener()
		listener.On("Accept").Return(nil, fmt.Errorf("some error"))

		// exercise

		ps, err := NewProxyService(
			connector,
			listener,
			logger,
			credentialRetriever.RetrieveCredentials)
		err = ps.Start()

		// sanity assert
		assert.NoError(t, err)
		if err != nil {
			return
		}

		// assert
		<-logger.ReceivedCall
		logger.AssertCalled(t, "Errorf")
	})

	t.Run("propagates error from connector", func(t *testing.T) {
		// prepare
		clientConn, _ := net.Pipe()

		logger := loggerMock.NewLogger()
		logger.On("Errorf").Return()

		connector := mock.NewConnector()
		connector.On("Connect").Return(nil, fmt.Errorf("some error"))

		credentialRetriever := mock.NewCredentialRetriever()
		credentialRetriever.On(
			"RetrieveCredentials",
		).Return(nil, nil)

		listener := mock.NewListener()
		listener.On("Accept").Return(clientConn, nil)

		// exercise

		ps, err := NewProxyService(
			connector,
			listener,
			logger,
			credentialRetriever.RetrieveCredentials)
		err = ps.Start()

		// sanity assert
		assert.NoError(t, err)
		if err != nil {
			return
		}

		// assert
		<-logger.ReceivedCall
		logger.AssertCalled(t, "Errorf")
	})

	t.Run("propagates error from retrieveCredentials", func(t *testing.T) {
		// prepare
		clientConn, _ := net.Pipe()
		backendConn, _ := net.Pipe()

		logger := loggerMock.NewLogger()
		logger.On("Errorf").Return()

		connector := mock.NewConnector()
		connector.On("Connect").Return(backendConn, nil)

		credentialRetriever := mock.NewCredentialRetriever()
		credentialRetriever.On(
			"RetrieveCredentials",
		).Return(nil, fmt.Errorf("some error"))

		listener := mock.NewListener()
		listener.On("Accept").Return(clientConn, nil)

		// exercise

		ps, err := NewProxyService(
			connector,
			listener,
			logger,
			credentialRetriever.RetrieveCredentials)
		err = ps.Start()

		// sanity assert
		assert.NoError(t, err)
		if err != nil {
			return
		}

		// assert
		<-logger.ReceivedCall
		logger.AssertCalled(t, "Errorf")
	})

	t.Run("proxy service streams packets in order from source to dest", func(t *testing.T) {
		// prepare

		// This allows us to control and view what happens in the client and
		// backend connections that Secretless is proxying.  Whatever we write
		// into `clientConnSrc` can be read from `clientConn` by the
		// ProxyService, and whatever the ProxyService writes into backendConn will be pipe into backendConnDest, so we can verify it.
		clientConn, clientConnSrc := net.Pipe()
		backendConn, backendConnDest := net.Pipe()

		logger := loggerMock.NewLogger()

		connector := mock.NewConnector()
		connector.On("Connect").Return(backendConn, nil)

		credentialRetriever := mock.NewCredentialRetriever()
		credentialRetriever.On("RetrieveCredentials").Return(nil, nil)

		listener := mock.NewListener()
		listener.On("Accept").Return(clientConn, nil)

		// exercise
		ps, err := NewProxyService(
			connector,
			listener,
			logger,
			credentialRetriever.RetrieveCredentials)

		err = ps.Start()
		// sanity check
		assert.NoError(t, err)
		if err != nil {
			return
		}

		go func() {
			_, _ = clientConnSrc.Write([]byte(testString1))
			_, _ = clientConnSrc.Write([]byte(testString2))
			_, _ = clientConnSrc.Write([]byte(testString3))
		}()

		// assert
		data := make([]byte, 256)

		// assert first packet
		dataLen, err := backendConnDest.Read(data)
		assert.Equal(
			t,
			string(data[:dataLen]),
			testString1,
		)

		// assert second packet
		dataLen, err = backendConnDest.Read(data)
		assert.Equal(
			t,
			string(data[:dataLen]),
			testString2,
		)

		// assert third packet
		dataLen, err = backendConnDest.Read(data)
		assert.Equal(
			t,
			string(data[:dataLen]),
			testString3,
		)
	})

	t.Run("proxy service streams packets between source and dest", func(t *testing.T) {
		// prepare
		clientConn, clientConnSrc := net.Pipe()
		backendConn, backendConnDest := net.Pipe()

		logger := loggerMock.NewLogger()

		connector := mock.NewConnector()
		connector.On("Connect").Return(backendConn, nil)

		credentialRetriever := mock.NewCredentialRetriever()
		credentialRetriever.On("RetrieveCredentials").Return(nil, nil)

		listener := mock.NewListener()
		listener.On("Accept").Return(clientConn, nil)

		// exercise
		ps, err := NewProxyService(
			connector,
			listener,
			logger,
			credentialRetriever.RetrieveCredentials)
		err = ps.Start()
		// sanity check
		assert.NoError(t, err)
		if err != nil {
			return
		}

		go func() {
			_, _ = clientConnSrc.Write([]byte(testString1))
			_, _ = backendConnDest.Write([]byte(testString2))
		}()

		// assert
		data := make([]byte, 256)

		// assert on client write
		dataLen, err := backendConnDest.Read(data)
		assert.Equal(
			t,
			string(data[:dataLen]),
			testString1,
		)

		// assert on backend write
		dataLen, err = clientConnSrc.Read(data)
		assert.Equal(
			t,
			string(data[:dataLen]),
			testString2,
		)
	})
}
