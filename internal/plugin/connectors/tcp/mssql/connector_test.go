package mssql
//
//import (
//	"context"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/mock"
//	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
//	logmock "github.com/cyberark/secretless-broker/pkg/secretless/log/mock"
//)
//
////type NewMSSQLConnectorFunc func(dsn string) (MSSQLConnector, error)
////type MSSQLConnector interface {
////  Connect(context.Context) (NetConner, error)
////}
////
//func TestHappyPath(t *testing.T) {
//	logger := logmock.NewLogger()
//	ctor := mock.NewSuccessfulMSSQLConnectorConstructor(
//		func(context.Context) (types.NetConner, error) {
//			// open up the channel, put shit on it
//			return nil, nil
//		},
//	)
//
//	NewSingleUseConnectorWithOptions(logger, ctor)
//	var err string
//	assert.Nil(t, err)
//}
