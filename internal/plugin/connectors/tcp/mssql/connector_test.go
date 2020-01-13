package mssql

import (
	"errors"
	"io"
	"net"
	"testing"

	errorspkg "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/mock"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	pluginconnector "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/denisenkom/go-mssqldb"
)

func runDefaultTestConnect(
	connector *SingleUseConnector,
) (net.Conn, error) {
	clientConn := mock.NewNetConn(nil)
	creds := pluginconnector.CredentialValuesByID{
		"credName": []byte("secret"),
	}

	return connector.Connect(clientConn, creds)
}

func TestHappyPath(t *testing.T) {
	expectedBackendConn := mock.NewNetConn(nil)

	connector := NewSingleUseConnectorWithOptions(
		mock.DefaultConnectorOptions(),
		mock.MSSQLConnectorCtor(
			func(options *mock.MSSQLConnectorCtorOptions) {
				options.BackendConn = expectedBackendConn
			},
		),
	)
	actualBackendConn, err := runDefaultTestConnect(connector)

	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, expectedBackendConn, actualBackendConn)
}

func TestReadLoginSucceeds(t *testing.T) {
	// login request returned from ReadLogin
	expectedLoginRequest := &mssql.LoginRequest{}
	expectedLoginRequest.Database = "test-database"
	expectedLoginRequest.AppName = "test-app-name"

	// actual login request available to Connect via context
	var actualLoginRequest *mssql.LoginRequest

	connector := NewSingleUseConnectorWithOptions(
		mock.DefaultConnectorOptions(),
		func(connectorOptions *types.ConnectorOptions) {
			connectorOptions.ReadLoginRequest = func(r io.ReadWriteCloser) (*mssql.LoginRequest, error) {
				return expectedLoginRequest, nil
			}
		},
		mock.MSSQLConnectorCtor(
			func(opt *mock.MSSQLConnectorCtorOptions) {
				opt.ClientLoginRequestPtr = &actualLoginRequest
			},
		),
	)

	_, _ = runDefaultTestConnect(connector)

	assert.Equal(t, actualLoginRequest, expectedLoginRequest)
}

func TestReadLoginFails(t *testing.T) {
	// error returned from ReadLogin
	expectedErr := errors.New("test error")

	connector := NewSingleUseConnectorWithOptions(
		mock.DefaultConnectorOptions(),
		mock.MSSQLConnectorCtor(
			func(opt *mock.MSSQLConnectorCtorOptions) {
				opt.Err = expectedErr
			},
		),
	)

	_, actualError := runDefaultTestConnect(connector)

	assert.Equal(t, expectedErr, errorspkg.Cause(actualError))
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
