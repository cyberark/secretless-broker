package mssql

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/mock"
	logmock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
	pluginConnector "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/denisenkom/go-mssqldb"
)

func TestHappyPath(t *testing.T) {
	logger := logmock.NewLogger()
	clientConn := mock.NewNetConn(nil)
	expectedBackendConn := mock.NewNetConn(nil)
	creds := pluginConnector.CredentialValuesByID{
		"credName": []byte("secret"),
	}

	ctor := mock.NewSuccessfulMSSQLConnectorCtor(
		func(ctx context.Context) (net.Conn, error) {
			interceptor := mssql.ConnectInterceptorFromContext(ctx)

			interceptor.ServerPreLoginResponse <- map[uint8][]byte{ 0: {0, 0} }

			<- interceptor.ClientLoginRequest

			interceptor.ServerLoginResponse <- &mssql.LoginResponse{}

			return expectedBackendConn, nil
		},
	)

	connector := NewSingleUseConnectorWithOptions(
		logger,
		ctor,
		mock.SuccessfulReadPreloginRequest,
		mock.SuccessfulWritePreloginResponse,
		mock.SuccessfulReadLoginRequest,
		mock.SuccessfulWriteLoginResponse,
		mock.SuccessfulWriteError,
		mock.FakeTdsBufferCtor,
	)
	actualBackendConn, err := connector.Connect(clientConn, creds)

	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, expectedBackendConn, actualBackendConn)
}

func TestWritePreloginSucceeds(t *testing.T) {
	logger := logmock.NewLogger()
	clientConn := mock.NewNetConn(nil)
	expectedBackendConn := mock.NewNetConn(nil)
	creds := pluginConnector.CredentialValuesByID{
		"credName": []byte("secret"),
	}

	// The fields we should get back from inside the mssql.Connect method
	fakePreloginResponse := map[uint8][]byte{
		mssql.PreloginVERSION:    {0, 0, 0, 0, 0, 0},
		mssql.PreloginENCRYPTION: {mssql.EncryptOn},
		mssql.PreloginINSTOPT:    {0},
		mssql.PreloginTHREADID:   {0, 0, 0, 0},
		mssql.PreloginMARS:       {0}, // MARS disabled
	}

	// The fields we should be sending to the client
	expectedPreloginResponse := map[uint8][]byte{
		mssql.PreloginVERSION:    {0, 0, 0, 0, 0, 0},
		mssql.PreloginENCRYPTION: {mssql.EncryptNotSup},
		mssql.PreloginINSTOPT:    {0},
		mssql.PreloginTHREADID:   {0, 0, 0, 0},
		mssql.PreloginMARS:       {0}, // MARS disabled
	}

	// The fields we actually receive after modifying them
	var actualPreloginResponse map[uint8][]byte

	// Expose our writePreloginMethod inside of Connect()
	mockWritePreloginResponse := func(_w io.ReadWriteCloser,
		fields map[uint8][]byte) error {
			actualPreloginResponse = fields
			return nil
	}

	ctor := mock.NewSuccessfulMSSQLConnectorCtor(
		func(ctx context.Context) (net.Conn, error) {
			interceptor := mssql.ConnectInterceptorFromContext(ctx)

			interceptor.ServerPreLoginResponse <- fakePreloginResponse

			<- interceptor.ClientLoginRequest

			interceptor.ServerLoginResponse <- &mssql.LoginResponse{}

			return expectedBackendConn, nil
		},
	)

	connector := NewSingleUseConnectorWithOptions(
		logger,
		ctor,
		mock.SuccessfulReadPreloginRequest,
		mockWritePreloginResponse,
		mock.SuccessfulReadLoginRequest,
		mock.SuccessfulWriteLoginResponse,
		mock.SuccessfulWriteError,
		mock.TdsBufferCtor,
	)

	connector.Connect(clientConn, creds)

	assert.Equal(t, expectedPreloginResponse, actualPreloginResponse)
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
