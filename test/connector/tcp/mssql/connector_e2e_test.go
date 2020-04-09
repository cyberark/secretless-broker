package mssqltest

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/cyberark/secretless-broker/internal/log"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp"
	secretlessMssql "github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	mssql "github.com/denisenkom/go-mssqldb"
)

func randomPortListener() (listener net.Listener, err error) {
	listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}

	// We don't want to wait forever for a connection to come in
	err = listener.(*net.TCPListener).SetDeadline(time.Now().Add(5 * time.Second))

	return
}

func exampleRun() (*TargetCapture, error) {
	logger := log.New(true)
	testLogger := logger.CopyWith("[TEST]", true)

	targetSvcListener, err := randomPortListener()
	if err != nil {
		return nil, err
	}
	targetSvcHost, targetSvcPort, err := net.SplitHostPort(
		targetSvcListener.Addr().String(),
		)
	if err != nil {
		return nil, err
	}

	secretlessSvcListener, err := randomPortListener()
	if err != nil {
		return nil, err
	}
	secretlessSvcHost, secretlessSvcPort, err := net.SplitHostPort(
		secretlessSvcListener.Addr().String(),
	)
	if err != nil {
		return nil, err
	}
	secretlessSvcPortInt, err := strconv.Atoi(secretlessSvcPort)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = targetSvcListener.Close()
	}()

	credentials := map[string][]byte{
		"sslmode": []byte("disable"),
		"host": []byte(targetSvcHost),
		"port": []byte(targetSvcPort),
	}

	defer func() {
		_ = secretlessSvcListener.Close()
	}()


	proxySvc, err := tcp.NewProxyService(
		secretlessMssql.NewConnector(
			connector.NewResources(nil, logger),
		),
		secretlessSvcListener,
		logger,
		func() (bytes map[string][]byte, e error) {
			return credentials, nil
		},
		)
	defer func() {
		_ = proxySvc.Stop()
	}()
	if err != nil {
		return nil, err
	}

	go func() {
		_ = proxySvc.Start()
	}()

	go func() {
		testLogger.Info("Client connecting to ", secretlessSvcHost, secretlessSvcPort)

		// we don't actually care about the response
		_, err = sqlcmdExec(
			dbConfig{
				Host:     secretlessSvcHost,
				Port:     secretlessSvcPortInt ,
				Username: "dummy",
				Password: "dummy",
				Database: "test-db",
				ReadOnly: true,
			},
			"",
		)
		if err != nil {
			testLogger.Infof("Failed to start client", err)
		}
	}()

	 capture, err := runMockTarget(targetSvcListener)
	 testLogger.Info("Captured: %v", capture)

	return capture, err
}


type TargetCapture struct {
	preloginRequest map[uint8][]byte
	loginRequest mssql.LoginRequest
}

func runMockTarget(listener net.Listener) (*TargetCapture, error) {
	targetCapture := &TargetCapture{}

	clientConnection, err := listener.Accept()
	if err != nil {
		return nil, err
	}

	// Set a deadline so that if things hang then they fail fast
	readWriteDeadline := time.Now().Add(5 * time.Second)
	err = clientConnection.SetDeadline(readWriteDeadline)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = clientConnection.Close()
	}()

	// Read prelogin request
	preloginRequest, err := mssql.ReadPreloginRequest(clientConnection)
	if err != nil {
		return targetCapture, err
	}

	targetCapture.preloginRequest = preloginRequest

	// Write prelogin response
	// ensuring no TLS support for now
	// TODO: add TLS handshake logic lol
	preloginResponse := preloginRequest
	preloginResponse[mssql.PreloginVERSION] = []byte{0x0e, 0x00, 0x0c, 0xa6, 0x00, 0x00}
	preloginResponse[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}
	err = mssql.WritePreloginResponse(clientConnection, preloginResponse)
	if err != nil {
		return targetCapture, err
	}

	// read login request
	loginRequest, err := mssql.ReadLoginRequest(clientConnection)
	if err != nil {
		return targetCapture, err
	}

	targetCapture.loginRequest = *loginRequest

	// write a dummy successful login response
	loginResponse := &mssql.LoginResponse{}
	loginResponse.ProgName = "test"
	loginResponse.TDSVersion = 0x730A0003
	loginResponse.Interface = 27
	err = mssql.WriteLoginResponse(clientConnection, loginResponse)
	if err != nil {
		return targetCapture, err
	}

	return targetCapture, nil
}

func TestConnectorE2E(t *testing.T) {
	_, _ = exampleRun()
}