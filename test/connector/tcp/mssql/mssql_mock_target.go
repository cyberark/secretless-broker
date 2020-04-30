package mssqltest

import (
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"

	mssql "github.com/denisenkom/go-mssqldb"
)

type mockTargetCapture struct {
	preloginRequest map[uint8][]byte
	loginRequest    mssql.LoginRequest
}

type mockTargetResult struct {
	capture *mockTargetCapture
	err     error
}

func newMockTarget(port string) (*mockTarget, error) {
	listener, err := ephemeralListenerOnPort(port)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(
		listener.Addr().String(),
	)
	if err != nil {
		return nil, err
	}

	return &mockTarget{
		listener: listener,
		host:     host,
		port:     port,
	}, nil
}

type mockTarget struct {
	listener  net.Listener
	host      string
	port      string
	accepting sync.Mutex
}

func (m *mockTarget) accept() chan mockTargetResult {
	m.accepting.Lock()
	mockTargetResponseChan := make(chan mockTargetResult)

	go func() {
		defer m.accepting.Unlock()

		clientConnection, err := m.listener.Accept()
		if err != nil {
			mockTargetResponseChan <- mockTargetResult{
				err: errors.Wrap(err, "mock target"),
			}
			return
		}

		capture, err := m.handleConnection(clientConnection)
		if err != nil {
			err = errors.Wrap(err, "mock target")
		}
		mockTargetResponseChan <- mockTargetResult{
			capture: capture,
			err:     err,
		}
	}()

	return mockTargetResponseChan
}

func (m *mockTarget) handleConnection(clientConnection net.Conn) (*mockTargetCapture, error) {
	var err error
	targetCapture := &mockTargetCapture{}

	// Set a deadline so that if things hang then they fail fast
	readWriteDeadline := time.Now().Add(1 * time.Second)
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

func (m *mockTarget) close() {
	_ = m.listener.Close()
}
