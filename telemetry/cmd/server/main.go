package main

import (
	"fmt"
	"io"
	"net"
	"os"

	mssql "github.com/denisenkom/go-mssqldb"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide port number")
		return
	}

	t, err := newMockTarget(arguments[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	t.Listen()
}

// mockTarget is a fake MSSQL server that can perform a plaintext handshake with an MSSQL
// client.
type mockTarget struct {
	listener   net.Listener
}

func newMockTarget(port string) (*mockTarget, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return nil, err
	}

	_, _, err = net.SplitHostPort(
		listener.Addr().String(),
	)
	if err != nil {
		return nil, err
	}

	return &mockTarget{
		listener: listener,
	}, nil
}

func (m *mockTarget) Listen() {
	for {
		clientConnection, err := m.listener.Accept()
		if err != nil {
			continue
		}

		go func() {
			fmt.Println("Starting handshake")
			err := m.handleHandshake(clientConnection)
			fmt.Println("Handshake done")
			if err != nil {
				return
			}

			// Return to sender :)
			fmt.Println("Echoing")
			io.CopyBuffer(clientConnection, clientConnection, make([]byte, 20))
			fmt.Println("Done")
		}()
	}
}

func (m *mockTarget) handleHandshake(clientConnection net.Conn) (error) {
	// Set a deadline so that if things hang then they fail fast
	//deadline := time.Now().Add(1 * time.Second)
	//err := clientConnection.SetDeadline(deadline)
	//if err != nil {
	//	return err
	//}

	//defer func() {
	//	_ = clientConnection.Close()
	//}()

	// Read prelogin request
	preloginRequest, err := mssql.ReadPreloginRequest(clientConnection)
	if err != nil {
		return err
	}

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
		return err
	}

	// Read login request from client
	_, err = mssql.ReadLoginRequest(clientConnection)
	if err != nil {
		return err
	}

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
		return err
	}

	return nil
}
