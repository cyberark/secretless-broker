package mssql

import (
	"context"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	"io"
	"net"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/mock"
	logmock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
	pluginconnector "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/denisenkom/go-mssqldb"
)

func defaultMSSQLConnectorCtor(expectedBackendConn *mock.NetConn) types.MSSQLConnectorCtor {
	return mock.NewSuccessfulMSSQLConnectorCtor(
		func(ctx context.Context) (net.Conn, error) {
			interceptor := mssql.ConnectInterceptorFromContext(ctx)

			interceptor.ServerPreLoginResponse <- map[uint8][]byte{0: {0, 0}}

			<-interceptor.ClientLoginRequest

			interceptor.ServerLoginResponse <- &mssql.LoginResponse{}

			return expectedBackendConn, nil
		},
	)
}

func defaultSingleUseConnector(expectedBackendConn *mock.NetConn) *SingleUseConnector {
	return NewSingleUseConnectorWithOptions(
		logmock.NewLogger(),
		defaultMSSQLConnectorCtor(expectedBackendConn),
		mock.SuccessfulReadPreloginRequest,
		mock.SuccessfulWritePreloginResponse,
		mock.SuccessfulReadLoginRequest,
		mock.SuccessfulWriteLoginResponse,
		mock.SuccessfulWriteError,
		mock.FakeTdsBufferCtor,
	)
}

func TestHappyPath(t *testing.T) {
	expectedBackendConn := mock.NewNetConn(nil)
	connector := defaultSingleUseConnector(expectedBackendConn)

	clientConn := mock.NewNetConn(nil)
	creds := pluginconnector.CredentialValuesByID{
		"credName": []byte("secret"),
	}
	actualBackendConn, err := connector.Connect(clientConn, creds)

	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, expectedBackendConn, actualBackendConn)
}

func TestFailingReadPrelogin(t *testing.T) {
	// Set up a default single use connector
	expectedBackendConn := mock.NewNetConn(nil)
	connector := defaultSingleUseConnector(expectedBackendConn)

	// Overwrite the readPreloginRequest function with one that returns a specific error.
	testError := "injected readPrelogin error"
	failingReadPreloginRequest := func(r io.ReadWriteCloser) (map[uint8][]byte, error) {
		return nil, errors.New(testError)
	}
	connector.readPreloginRequest = failingReadPreloginRequest
	connector.writeError = mock.WriteError

	clientConn := mock.NewNetConn(nil)
	creds := pluginconnector.CredentialValuesByID{
		"credName": []byte("secret"),
	}
	_, err := connector.Connect(clientConn, creds)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), testError)

	// Check that expected error was written to the client.
	output := clientConn.WriteHistory[0]
	assert.Contains(t, string(output), testError)
}

/*
Test cases not to forget

	- This is not exhaustive.  Use the method of tracing all the code paths (for
	each error condition, assuming the previous succeeded) and add a test for
	each.  If that becomes too many, use judgment to eliminate less important
	ones.

	- While we shouldn't test the logger messages extensively, here is one that
	we should test: sending an error message to the user fails.  We want to make
	sure those are logged.
*/
