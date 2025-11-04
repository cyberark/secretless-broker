package mysql

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mysql/protocol"
	conplugin "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

type dummyAddr string

func (d dummyAddr) Network() string { return "tcp" }
func (d dummyAddr) String() string  { return string(d) }

type mockConn struct {
	writes     [][]byte
	closed     bool
	writeError error
}

func (m *mockConn) Read(b []byte) (int, error) { return 0, io.EOF }
func (m *mockConn) Write(b []byte) (int, error) {
	if m.writeError != nil {
		return 0, m.writeError
	}
	m.writes = append(m.writes, append([]byte(nil), b...))
	return len(b), nil
}
func (m *mockConn) Close() error                     { m.closed = true; return nil }
func (m *mockConn) LocalAddr() net.Addr              { return dummyAddr("local") }
func (m *mockConn) RemoteAddr() net.Addr             { return dummyAddr("remote") }
func (m *mockConn) SetDeadline(time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(time.Time) error { return nil }

func totalWritten(m *mockConn) int {
	total := 0
	for _, w := range m.writes {
		total += len(w)
	}
	return total
}

func combinedWrites(m *mockConn) []byte {
	var out []byte
	for _, w := range m.writes {
		out = append(out, w...)
	}
	return out
}

func TestSendErrorToClient_WithProtocolError_WritesPacket(t *testing.T) {
	client := &mockConn{}
	sut := &SingleUseConnector{
		mySQLClientConn: NewClientConnection(client),
	}

	sut.sendErrorToClient(protocol.NewGenericError(errors.New("boom")))

	if totalWritten(client) == 0 {
		t.Fatalf("expected bytes written to client, got none")
	}
}

func TestSendErrorToClient_WithGenericError_WritesPacket(t *testing.T) {
	client := &mockConn{}
	sut := &SingleUseConnector{
		mySQLClientConn: NewClientConnection(client),
	}

	sut.sendErrorToClient(errors.New("generic"))

	if totalWritten(client) == 0 {
		t.Fatalf("expected bytes written to client, got none")
	}
}

func TestSendErrorToClient_GenericError_WritesExpectedPacket(t *testing.T) {
	client := &mockConn{}
	sut := &SingleUseConnector{
		mySQLClientConn: NewClientConnection(client),
	}

	inputErr := errors.New("some error")
	expected := protocol.NewGenericError(inputErr).GetPacket()

	sut.sendErrorToClient(inputErr)

	got := combinedWrites(client)
	if !bytes.Equal(got, expected) {
		t.Fatalf("expected packet %v, got %v", expected, got)
	}
}

func TestConnect_MissingRequiredCredentials_SendsErrorAndReturnsError(t *testing.T) {
	client := &mockConn{}
	sut := &SingleUseConnector{}

	creds := conplugin.CredentialValuesByID{
		"host": []byte("localhost"),
		"port": []byte("3306"),
	}

	conn, err := sut.Connect(client, creds)

	if err == nil {
		t.Fatalf("expected error for missing required credentials, got nil")
	}
	if conn != nil {
		t.Fatalf("expected nil backend connection, got non-nil")
	}
	if totalWritten(client) == 0 {
		t.Fatalf("expected error packet written to client, got none")
	}
}

func TestConnect_DialBackendError_SendsErrorAndReturnsError(t *testing.T) {
	client := &mockConn{}
	sut := &SingleUseConnector{}

	// Port 1 on localhost should refuse immediately.
	creds := conplugin.CredentialValuesByID{
		"host":     []byte("127.0.0.1"),
		"port":     []byte("1"),
		"username": []byte("u"),
		"password": []byte("p"),
	}

	conn, err := sut.Connect(client, creds)

	if err == nil {
		t.Fatalf("expected dial error, got nil")
	}
	if conn != nil {
		t.Fatalf("expected nil backend connection, got non-nil")
	}
	if totalWritten(client) == 0 {
		t.Fatalf("expected error packet written to client, got none")
	}
}

func TestConnect_HandshakeErrorAfterDial_SendsErrorAndReturnsError(t *testing.T) {
	// Start a dummy backend that accepts and closes immediately to force handshake failure.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	done := make(chan struct{})
	go func() {
		if c, e := ln.Accept(); e == nil && c != nil {
			_ = c.Close()
		}
		close(done)
	}()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())

	client := &mockConn{}
	sut := &SingleUseConnector{}

	creds := conplugin.CredentialValuesByID{
		"host":     []byte("127.0.0.1"),
		"port":     []byte(portStr),
		"username": []byte("u"),
		"password": []byte("p"),
	}

	conn, err := sut.Connect(client, creds)

	<-done

	if err == nil {
		t.Fatalf("expected handshake error, got nil")
	}
	if conn != nil {
		t.Fatalf("expected nil backend connection, got non-nil")
	}
	if totalWritten(client) == 0 {
		t.Fatalf("expected error packet written to client, got none")
	}
}
