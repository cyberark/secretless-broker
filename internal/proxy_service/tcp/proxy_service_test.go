package tcp

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/proxy_service/mock"
)

const testString1 = "good heavens"
const testString2 = "moderate heavens"

func TestNewProxyService(t *testing.T) {
    t.Run("empty constructor arguments result in errors", func(t *testing.T) {
        _, err := NewProxyService(
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
            mock.NewConnector().Connect,
			mock.NewCredentialRetriever().RetrieveCredentials,
            mock.NewListener(),
        )
        assert.NoError(t, err)
        if err != nil {
            return
        }
    })

    // TODO: propagates error from retrieveCredentials

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
            connector:           connector.Connect,
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
            string(creds["credName"]),
            strings.Repeat("\x00", len([]byte("credValue"))),
        )
    })
}

func TestProxyService_Start(t *testing.T) {
    t.Run("stopped proxy service cannot be restarted", func(t *testing.T) {
		connector := mock.NewConnector()
		credentialRetriever := mock.NewCredentialRetriever()
		listener := mock.NewListener()
		listener.On("Close").Return(nil)

		ps, err := NewProxyService(
			connector.Connect,
			credentialRetriever.RetrieveCredentials,
			listener)

		err = ps.Stop()
		assert.NoError(t, err)
		if err != nil {
			return
		}

		err = ps.Start()
		assert.Error(t, err)
		if err == nil {
			return
		}
	})

    t.Run("proxy service streams from source to dest", func(t *testing.T) {
    	// prepare

		// This allows us to control and view what happens in the client and
		// backend connections that Secretless is proxying.  Whatever we write
		// into `clientConnSrc` can be read from `clientConn` by the
		// ProxyService, and whatever the ProxyService writes into backendConn will be pipe into backendConnDest, so we can verify it.
        clientConn, clientConnSrc := net.Pipe()
		backendConn, backendConnDest := net.Pipe()

		connector := mock.NewConnector()
		connector.On("Connect").Return(
			backendConn,
			nil)

		credentialRetriever := mock.NewCredentialRetriever()
		credentialRetriever.On("RetrieveCredentials").Return(
			nil,
			nil)

		listener := mock.NewListener()
		listener.On("Accept").Return(clientConn, nil)

        // exercise

        // tcp service with 'backendConn" mock and listener whose first call to Accept return the
        // "clientConn" mock then blocks forever
        ps, err := NewProxyService(
            connector.Connect,
            credentialRetriever.RetrieveCredentials,
            listener)
        err = ps.Start()

        // sanity assert
        assert.NoError(t, err)
        if err != nil {
            return
        }

        go clientConnSrc.Write([]byte(testString1))

        data := make([]byte, 256)
        dataLen, err := backendConnDest.Read(data)

        // assert
        assert.NoError(t, err)
        if err != nil {
            return
        }
        assert.Equal(
            t,
            string(data[:dataLen]),
            testString1,
        )

		go clientConnSrc.Write([]byte(testString2))

		dataLen, err = backendConnDest.Read(data)

		// assert
		assert.NoError(t, err)
		if err != nil {
			return
		}
		assert.Equal(
			t,
			string(data[:dataLen]),
			testString2,
		)
    })
}
