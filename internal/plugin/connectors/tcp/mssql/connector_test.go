package mssql

//import (
//	"context"
//	"fmt"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/mock"
//	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
//	logmock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
//	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
//	"github.com/cyberark/secretless-broker/third_party/ctxtypes"
//)
//
//func TestHappyPath(t *testing.T) {
//	logger := logmock.NewLogger()
//	clientConn := mock.NewNetConn()
//	expectedBackendConn := mock.NewNetConn()
//	creds := connector.CredentialValuesByID{
//		"credName": []byte("secret"),
//	}
//
//	ctor := mock.NewSuccessfulMSSQLConnectorConstructor(
//		func(ctx context.Context) (types.NetConner, error) {
//			fmt.Println("Hey 1")
//			preLoginResponse := ctx.Value(ctxtypes.PreLoginResponseKey).(chan map[uint8][]byte)
//			preLoginResponse <- map[uint8][]byte{ 0: {0, 0} }
//			fmt.Println("Hey 2")
//			return expectedBackendConn, nil
//		},
//	)
//
//	connector := NewSingleUseConnectorWithOptions(
//		logger,
//		ctor,
//		mock.SuccessfulReadPrelogin,
//		mock.SuccessfulWritePrelogin,
//	)
//	actualBackendConn, err := connector.Connect(clientConn, creds)
//
//	assert.Nil(t, err)
//	if err != nil {
//		return
//	}
//
//	assert.Equal(t, expectedBackendConn, actualBackendConn)
//}
