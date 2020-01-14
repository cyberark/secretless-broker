package mssql

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql/types"
	mssql "github.com/denisenkom/go-mssqldb"
)

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
