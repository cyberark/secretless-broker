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
	mssql "github.com/denisenkom/go-mssqldb"
)

func TestSingleUseConnector_Connect(t *testing.T) {
	t.Run("singleUseConnector.driver#Connect success", func(t *testing.T) {
		expectedBackendConn := mock.DefaultMSSQLConnectorCtor.BackendConn

		connector := newSingleUseConnectorWithOptions(
			mock.DefaultConnectorOptions,
		)
		actualBackendConn, err := runDefaultTestConnect(connector)

		assert.Nil(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, expectedBackendConn, actualBackendConn)
	})

	t.Run("singleUseConnector.driver#Connect fail", func(t *testing.T) {
		var methodFailsExpectedErr = errors.New("failed to complete connection")

		customCtor := mock.DefaultMSSQLConnectorCtor
		customCtor.Err = methodFailsExpectedErr

		methodFails(t, methodFailsExpectedErr, mock.CustomNewMSSQLConnectorOption(customCtor))
	})

	t.Run("singleUseConnector#ReadPreLoginRequest fail", func(t *testing.T) {
		var methodFailsExpectedErr = errors.New("failed to read prelogin request from client")

		methodFails(t, methodFailsExpectedErr, func(connectorOptions *types.ConnectorOptions) {
			connectorOptions.ReadPreloginRequest = func(
				io.ReadWriteCloser,
			) (map[uint8][]byte, error) {
				return nil, methodFailsExpectedErr
			}
		})
	})

	t.Run("singleUseConnector#WritePreloginResponse succeeds", func(t *testing.T) {

		// The fields we should get back from inside the mssql.Connect method
		fakePreLoginResponse := map[uint8][]byte{
			mssql.PreloginVERSION:    {0, 0, 0, 0, 0, 0},
			mssql.PreloginENCRYPTION: {mssql.EncryptOn},
			mssql.PreloginINSTOPT:    {0},
			mssql.PreloginTHREADID:   {0, 0, 0, 0},
			mssql.PreloginMARS:       {0}, // MARS disabled
		}

		// The fields we should be sending to the client
		expectedPreLoginResponse := map[uint8][]byte{
			mssql.PreloginVERSION:    {0, 0, 0, 0, 0, 0},
			mssql.PreloginENCRYPTION: {mssql.EncryptNotSup},
			mssql.PreloginINSTOPT:    {0},
			mssql.PreloginTHREADID:   {0, 0, 0, 0},
			mssql.PreloginMARS:       {0}, // MARS disabled
		}

		// The fields we actually receive after modifying them
		var actualPreLoginResponse map[uint8][]byte

		var actualClient io.ReadWriteCloser

		customCtor := mock.DefaultMSSQLConnectorCtor
		customCtor.ServerPreloginResponse = fakePreLoginResponse

		connector := newSingleUseConnectorWithOptions(
			mock.DefaultConnectorOptions,
			mock.CustomNewMSSQLConnectorOption(customCtor),
			func(connectorOptions *types.ConnectorOptions) {
				connectorOptions.WritePreloginResponse = func(
					w io.ReadWriteCloser,
					fields map[uint8][]byte,
				) error {
					actualClient = w
					actualPreLoginResponse = fields
					return nil
				}
			},
		)

		_, _ = runDefaultTestConnect(connector)
		expectedClient := connector.clientConn

		assert.Equal(t, actualPreLoginResponse, expectedPreLoginResponse)
		// confirm that WritePreloginResponse is called with the client connection
		assert.Equal(t, expectedClient, actualClient)
	})

	t.Run("singleUseConnector#WritePreloginResponse fails", func(t *testing.T) {
		var methodFailsExpectedErr = errors.New("failed to write prelogin response to client")

		methodFails(t, methodFailsExpectedErr, func(connectorOptions *types.ConnectorOptions) {
			connectorOptions.WritePreloginResponse = func(
				io.ReadWriteCloser,
				map[uint8][]byte,
			) error {
				return methodFailsExpectedErr
			}
		})
	})

	t.Run("singleUseConnector#ReadLoginRequest succeeds", func(t *testing.T) {
		// expected login request returned from ReadLoginRequest
		expectedLoginRequest := &mssql.LoginRequest{}
		expectedLoginRequest.Database = "test-database"
		expectedLoginRequest.AppName = "test-app-name"

		// actual login request sent to Connect via context
		var actualLoginRequest *mssql.LoginRequest
		var actualClient io.ReadWriteCloser

		customCtor := mock.DefaultMSSQLConnectorCtor
		customCtor.ClientLoginRequestPtr = &actualLoginRequest

		connector := newSingleUseConnectorWithOptions(
			mock.DefaultConnectorOptions,
			mock.CustomNewMSSQLConnectorOption(customCtor),
			func(connectorOptions *types.ConnectorOptions) {
				connectorOptions.ReadLoginRequest = func(r io.ReadWriteCloser) (*mssql.LoginRequest, error) {
					actualClient = r
					return expectedLoginRequest, nil
				}
			},
		)

		_, _ = runDefaultTestConnect(connector)
		expectedClient := connector.clientConn

		assert.Equal(t, actualLoginRequest, expectedLoginRequest)
		// confirm that ReadLoginRequest is called with the client connection
		assert.Equal(t, expectedClient, actualClient)
	})

	t.Run("singleUseConnector#ReadLoginRequest fail", func(t *testing.T) {
		var methodFailsExpectedErr = errors.New("failed to handle login from client")

		methodFails(t, methodFailsExpectedErr, func(connectorOptions *types.ConnectorOptions) {
			connectorOptions.ReadLoginRequest = func(
				r io.ReadWriteCloser,
			) (request *mssql.LoginRequest, e error) {
				return nil, methodFailsExpectedErr
			}
		})
	})
}

// Test helpers
//

// newSingleUseConnectorWithOptions creates a new SingleUseConnector, and allows
// you to specify the newMSSQLConnector explicitly.  Intended to be used in unit
// tests only.
func newSingleUseConnectorWithOptions(
	options ...types.ConnectorOption,
) *SingleUseConnector {
	// Default Options
	args := &types.ConnectorOptions{}

	for _, option := range options {
		option(args)
	}

	return &SingleUseConnector{
		ConnectorOptions: *args,
	}
}

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

// methodFails checks that the expected error is present in:
// 1. the error returned from the call to the Connect method
// 2. the error propagated to the client
func methodFails(
	t *testing.T,
	methodFailsExpectedErr error,
	connectorOptions types.ConnectorOption,
) {
	var actualClientErr error
	var actualClient io.ReadWriteCloser
	var actualErr error

	// expected error on method
	expectedErr := methodFailsExpectedErr

	// We build our connector in three parts
	// 1. Pass in our Mock connector with default methods
	// 2. Pass in our error writer for this test case
	// 3. Pass in our failing method for this test case
	connector := newSingleUseConnectorWithOptions(
		mock.DefaultConnectorOptions,
		func(connectorOptions *types.ConnectorOptions) {
			// error should always be written to client
			connectorOptions.WriteError = func(w io.ReadWriteCloser, err mssql.Error) error {
				actualClient = w
				actualClientErr = err
				return nil
			}
		},
		connectorOptions,
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
