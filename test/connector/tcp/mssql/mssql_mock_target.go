package mssqltest

import (
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"

	mssql "github.com/denisenkom/go-mssqldb"
)

// mockTargetCapture is a capture of the packets from a client involved in a handshake
// with an MSSQL mock server (mockTarget).
type mockTargetCapture struct {
	preloginRequest map[uint8][]byte
	loginRequest    mssql.LoginRequest
}

type mockTargetResult struct {
	capture *mockTargetCapture
	err     error
}

// mockTarget is a fake MSSQL server that can perform a plaintext handshake with an MSSQL
// client. It's used to capture packets from the client and make assertions on them. It is
// particularly useful for ensuring relevant client parameters are propagated to the MSSQL
// server when Secretless is used to proxy a connection to an MSSQL server.
type mockTarget struct {
	listener   net.Listener
	host       string
	port       string
	acceptLock sync.Mutex
}

func newMockTarget(port string) (*mockTarget, error) {
	listener, err := localListenerOnPort(port)
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

// singleAcceptAndHandle handles a client connection on the mock target's listener. The
// mock target is only capable of accepting and handling a single connection. This is to
// allow coordination with test clients so that we know which client the listener has just
// accepted; without this we'd have no way of telling which client connection is
// currently being handled if multiple client connections are made at the same time.
func (m *mockTarget) singleAcceptAndHandle() chan mockTargetResult {
	m.acceptLock.Lock()
	mockTargetResponseChan := make(chan mockTargetResult)

	go func() {
		defer m.acceptLock.Unlock()

		// We generally don't want to wait forever for a connection to come in
		_ = m.listener.(*net.TCPListener).SetDeadline(time.Now().Add(2 * time.Second))

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
	targetCapture := &mockTargetCapture{}

	// Set a deadline so that if things hang then they fail fast
	deadline := time.Now().Add(1 * time.Second)
	err := clientConnection.SetDeadline(deadline)
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

	// Write prelogin response.
	// The PRELOGIN packet type is used for both the client request and server response.
	// It is described in detail at https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-tds/60f56408-0188-4cd5-8b90-25c6f2423868.

	// Craft the prelogin response from the prelogin request
	preloginResponse := preloginRequest
	// The version set here is taken from intercepted traffic over wireshark from an
	// actual MSSQL server.
	preloginResponse[mssql.PreloginVERSION] = []byte{0x0e, 0x00, 0x0c, 0xa6, 0x00, 0x00}
	// Ensure no TLS support, otherwise the client might try to upgrade
	preloginResponse[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}

	// Write prelogin response to client
	err = mssql.WritePreloginResponse(clientConnection, preloginResponse)
	if err != nil {
		return targetCapture, err
	}

	// Read login request from client
	loginRequest, err := mssql.ReadLoginRequest(clientConnection)
	if err != nil {
		return targetCapture, err
	}

	targetCapture.loginRequest = *loginRequest

	// Write a dummy successful login response to the client.
	// In the TDS protocol a login response is represented by a LOGINACK packet,
	// for details on this packet type see https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-tds/490e563d-cc6e-4c86-bb95-ef0186b98032.
	loginResponse := &mssql.LoginResponse{}
	// The name of the server.
	loginResponse.ProgName = "test"
	// The TDS version being used by the server.
	loginResponse.TDSVersion = 0x74000004 // TDS74
	// The type of interface with which the server will singleAcceptAndHandle client requests.
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
