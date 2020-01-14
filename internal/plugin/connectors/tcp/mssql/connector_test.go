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

func TestSingleUseConnector_Connect(t *testing.T) {
	t.Run("singleUseConnector.driver#Connect success", func(t *testing.T) {
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
	})

	t.Run("singleUseConnector.driver#Connect fail", func(t *testing.T) {
		methodFails(t, mock.MSSQLConnectorCtor(
			func(ctorOptions *mock.MSSQLConnectorCtorOptions) {
				ctorOptions.Err = methodFailsExpectedErr
			}),
		)
	})
}
// Test helpers
//

// runDefaultTestConnect passes in default values to the Connect method of a connector.
// This helps avoid boilerplate
func runDefaultTestConnect(
	connector *SingleUseConnector,
) (net.Conn, error) {
	clientConn := mock.NewNetConn(nil)
	creds := pluginconnector.CredentialValuesByID{
		"credName": []byte("secret"),
	}

	return connector.Connect(clientConn, creds)
}

// methodFailsExpectedErr is the error value used inside methodFails
var methodFailsExpectedErr = errors.New("failed for the test")

// methodFails checks that the expected error is present in:
// 1. the error returned from the call to the Connect method
// 2. the error propagated to the client
func methodFails(
	t *testing.T,
	connectorOption types.ConnectorOption,
) {
	var actualClientErr error
	var actualClient io.ReadWriteCloser
	var actualErr error

	// expected error on method
	expectedErr := methodFailsExpectedErr

	connector := NewSingleUseConnectorWithOptions(
		mock.DefaultConnectorOptions(),
		func(connectorOptions *types.ConnectorOptions) {
			// error should always be written to client
			connectorOptions.WriteError = func(w io.ReadWriteCloser, err mssql.Error) error {
				actualClient = w
				actualClientErr = err
				return nil
			}
		},
		connectorOption,
	)

	_, actualErr = runDefaultTestConnect(connector)
	expectedClient := connector.clientConn

	// confirms error returned by #Connect contains expected error
	assert.Equal(t, expectedErr, errorspkg.Cause(actualErr))
	// confirms that errors are always written to the client and not something else
	assert.Equal(t, expectedClient, actualClient)
	// confirms error written to client contains expected error
	// actualClientErr can be anything but it should contain the expected error
	assert.Contains(t, actualClientErr.Error(), expectedErr.Error())
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
