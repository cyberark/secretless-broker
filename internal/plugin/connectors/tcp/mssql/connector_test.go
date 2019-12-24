package mssql

import (
	"context"
	mssql "github.com/denisenkom/go-mssqldb"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/mock"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	logmock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

func TestHappyPath(t *testing.T) {
	logger := logmock.NewLogger()
	clientConn := mock.NewNetConn(nil)
	expectedBackendConn := mock.NewNetConn(nil)
	creds := connector.CredentialValuesByID{
		"credName": []byte("secret"),
	}

	ctor := mock.NewSuccessfulMSSQLConnectorCtor(
		func(ctx context.Context) (types.NetConner, error) {
			preLoginResponse := ctx.Value(mssql.ConnectInterceptorKey).(chan map[uint8][]byte)
			preLoginResponse <- map[uint8][]byte{0: {0, 0}}
			return expectedBackendConn, nil
		},
	)

	connector := NewSingleUseConnectorWithOptions(
		logger,
		ctor,
		mock.SuccessfulReadPrelogin,
		mock.SuccessfulWritePrelogin,
		mock.SuccessfulTdsBufferCtor(),
	)
	actualBackendConn, err := connector.Connect(clientConn, creds)

	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, expectedBackendConn, actualBackendConn)
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
