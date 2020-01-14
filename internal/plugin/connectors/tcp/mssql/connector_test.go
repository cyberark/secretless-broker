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

func TestThirdPartConnectSuccess(t *testing.T) {
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

func TestThirdPartConnectFail(t *testing.T) {
	methodFails(t, mock.MSSQLConnectorCtor(
		func(ctorOptions *mock.MSSQLConnectorCtorOptions) {
			ctorOptions.Err = methodFailsExpectedErr
		}),
	)
}

func TestProductionReadPreLoginRequest(t *testing.T) {
	// production version of ReadPreLoginRequest
	var readPreLoginRequest types.ReadPreloginRequestFunc = mssql.ReadPreloginRequest

	// expected prelogin request available from net.Conn passed to ReadLogin
	expectedPreLoginRequest := map[uint8][]byte{
		1: {2,3,4},
	}

	r, w := net.Pipe()
	go func() {
		_ = mssql.WritePreloginRequest(w, expectedPreLoginRequest)
	}()

	// prelogin request returned from ReadPreLoginRequest
	actualLoginRequest, _ := readPreLoginRequest(r)

	assert.Equal(t, actualLoginRequest, expectedPreLoginRequest)
}

func TestReadPreLoginRequestFails(t *testing.T) {
	methodFails(t, func(connectorOptions *types.ConnectorOptions) {
		connectorOptions.ReadPreloginRequest = func(
			io.ReadWriteCloser,
		) (map[uint8][]byte, error) {
			return nil, methodFailsExpectedErr
		}
	})
}

func TestProductionWritePreLoginResponse(t *testing.T) {
	// production version of writePreLoginResponse
	var writePreLoginResponse types.WritePreloginResponseFunc = mssql.WritePreloginResponse

	// expected prelogin request available from net.Conn passed to writePreLoginResponse
	expectedPreLoginResponse := map[uint8][]byte{
		1: {2,3,4},
	}

	r, w := net.Pipe()
	go func() {
		writePreLoginResponse(w, expectedPreLoginResponse)
	}()

	// prelogin response returned from ReadPreloginResponse
	actualPreLoginResponse, err := mssql.ReadPreloginResponse(r)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, expectedPreLoginResponse, actualPreLoginResponse)
}

func TestWritePreLoginResponseArgs(t *testing.T) {
	// expected prelogin response returned from server
	expectedPreLoginResponse := map[uint8][]byte {
		1: {2,3,4},
	}

	// actual prelogin response passed as args to WritePreloginResponse
	var actualPreLoginResponse map[uint8][]byte
	var actualClient io.ReadWriteCloser

	connector := NewSingleUseConnectorWithOptions(
		mock.DefaultConnectorOptions(),
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
		mock.MSSQLConnectorCtor(
			func(opt *mock.MSSQLConnectorCtorOptions) {
				opt.ServerPreloginResponse = expectedPreLoginResponse
			},
		),
	)

	_, _ = runDefaultTestConnect(connector)
	expectedClient := connector.clientConn

	assert.Equal(t, actualPreLoginResponse, expectedPreLoginResponse)
	// confirm that WritePreloginResponse is called with the client connection
	assert.Equal(t, expectedClient, actualClient)
}

func TestWritePreLoginResponseFails(t *testing.T) {
	methodFails(t, func(connectorOptions *types.ConnectorOptions) {
		connectorOptions.WritePreloginResponse = func(
			io.ReadWriteCloser,
			map[uint8][]byte,
		) error {
			return methodFailsExpectedErr
		}
	})
}

func TestProductionReadLoginRequest(t *testing.T) {
	// production version of ReadLoginRequest
	var readLoginRequest types.ReadLoginRequestFunc = mssql.ReadLoginRequest

	// expected login request available from net.Conn passed to ReadLogin
	expectedLoginRequest := &mssql.LoginRequest{}
	expectedLoginRequest.AppName = "test-app-name"
	expectedLoginRequest.UserName = "test-user-name"
	expectedLoginRequest.Database = "test-database"
	expectedLoginRequest.SSPI = []uint8{}

	r, w := net.Pipe()
	go func() {
		_ = mssql.WriteLoginRequest(w, expectedLoginRequest)
	}()

	// login request returned from ReadLoginRequest
	actualLoginRequest, _ := readLoginRequest(r)

	assert.Equal(t, actualLoginRequest, expectedLoginRequest)
}

func TestReadLoginRequestSucceeds(t *testing.T) {
	// expected login request returned from ReadLoginRequest
	expectedLoginRequest := &mssql.LoginRequest{}
	expectedLoginRequest.Database = "test-database"
	expectedLoginRequest.AppName = "test-app-name"

	// actual login request sent to Connect via context
	var actualLoginRequest *mssql.LoginRequest
	var actualClient io.ReadWriteCloser

	connector := NewSingleUseConnectorWithOptions(
		mock.DefaultConnectorOptions(),
		func(connectorOptions *types.ConnectorOptions) {
			connectorOptions.ReadLoginRequest = func(r io.ReadWriteCloser) (*mssql.LoginRequest, error) {
				actualClient = r
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
	expectedClient := connector.clientConn

	assert.Equal(t, actualLoginRequest, expectedLoginRequest)
	// confirm that ReadLoginRequest is called with the client connection
	assert.Equal(t, expectedClient, actualClient)
}

func TestReadLoginRequestFails(t *testing.T) {
	methodFails(t, func(connectorOptions *types.ConnectorOptions) {
		connectorOptions.ReadLoginRequest = func(
			r io.ReadWriteCloser,
		) (request *mssql.LoginRequest, e error) {
			return nil, methodFailsExpectedErr
		}
	})
}

func TestProductionWriteLoginResponse(t *testing.T) {
	// production version of WriteLoginResponse
	var writeLoginResponse types.WriteLoginResponseFunc = mssql.WriteLoginResponse

	// expected login request available from net.Conn passed to WriteLoginResponse
	expectedLoginResponse := &mssql.LoginResponse{}
	expectedLoginResponse.Interface = 23
	expectedLoginResponse.ProgName = "test-progname"
	expectedLoginResponse.ProgVer = 01
	expectedLoginResponse.TDSVersion = 12

	r, w := net.Pipe()
	go func() {
		writeLoginResponse(w, expectedLoginResponse)
	}()

	// login response returned from ReadLoginResponse
	actualLoginResponse, err := mssql.ReadLoginResponse(r)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, expectedLoginResponse, actualLoginResponse)
}

func TestWriteLoginResponseArgs(t *testing.T) {
	// expected login response returned from server
	expectedLoginResponse := &mssql.LoginResponse{}
	expectedLoginResponse.Interface = 23
	expectedLoginResponse.ProgName = "test-progname"
	expectedLoginResponse.ProgVer = 01
	expectedLoginResponse.TDSVersion = 12

	// actual login response passed as args to WriteLoginResponse
	var actualLoginResponse *mssql.LoginResponse
	var actualClient io.ReadWriteCloser

	connector := NewSingleUseConnectorWithOptions(
		mock.DefaultConnectorOptions(),
		func(connectorOptions *types.ConnectorOptions) {
			connectorOptions.WriteLoginResponse = func(
				w io.ReadWriteCloser,
				loginRes *mssql.LoginResponse,
			) error {
				actualClient = w
				actualLoginResponse = loginRes
				return nil
			}
		},
		mock.MSSQLConnectorCtor(
			func(opt *mock.MSSQLConnectorCtorOptions) {
				opt.ServerLoginResponse = expectedLoginResponse
			},
		),
	)

	_, _ = runDefaultTestConnect(connector)
	expectedClient := connector.clientConn

	assert.Equal(t, actualLoginResponse, expectedLoginResponse)
	// confirm that WriteLoginResponse is called with the client connection
	assert.Equal(t, expectedClient, actualClient)
}

func TestWriteLoginResponseFails(t *testing.T) {
	methodFails(t, func(connectorOptions *types.ConnectorOptions) {
		connectorOptions.WriteLoginResponse = func(
			w io.ReadWriteCloser,
			loginRes *mssql.LoginResponse,
		) error {
			return methodFailsExpectedErr
		}
	})
}

func TestProductionWriteErr(t *testing.T) {
	// production version of WriteError
	var writeError types.WriteErrorFunc = mssql.WriteError72

	// expected error
	expectedErr := mssql.Error{
		Number:     1,
		State:      2,
		Class:      3,
		Message:    "test-message",
		ServerName: "test-server-name",
		ProcName:   "test-proc-name",
		LineNo:     4,
	}

	r, w := net.Pipe()
	go func() {
		writeError(w, expectedErr)
	}()

	// production read error
	actualErr, err := mssql.ReadError(r)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Contains(t, actualErr.Error(), expectedErr.Error())
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
