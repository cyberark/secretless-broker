package tcp

import (
	"fmt"
	"io"
	"net"

	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// Zeroizes the values of the fetched credentials. We don't want to
// rely on garbage collection for this (it might be slow and/or only free them) so
// we manually clear
func zeroizeCredentials(backendCredentials map[string][]byte) {
	for _, credentialBytes := range backendCredentials {
		for i := range credentialBytes {
			credentialBytes[i] = 0
		}
	}
}

func duplexStream(source io.ReadWriter, destination io.ReadWriter) <-chan error {
	errChan := make(chan error)
	go func() {
		errChan <- stream(source, destination)
	}()
	go func() {
		errChan <- stream(destination, source)
	}()
	return errChan
}

func stream(source io.Reader, destination io.Writer) error {
	for {
		if _, err := io.Copy(destination, source); err != nil {
			return err
		}
		return nil
	}
}

type proxyService struct {
	connector           tcp.Connector
	retrieveCredentials internal.CredentialsRetriever
	listener            net.Listener
	done                bool
	error               chan error
}

// NewProxyService constructs a new instance of a TCP ProxyService. The
// constructor takes a TCP Connector, CredentialResolver and Listener.
// A TCP ProxyService is able to Connect with Credentials then subsequently stream
// bytes between client and target service
func NewProxyService(
	connector tcp.Connector,
	resolveCredentials internal.CredentialsRetriever,
	listener net.Listener,
) (internal.ProxyService, error) {
	errors := validation.Errors{}

	if connector == nil {
		errors["connector"] = fmt.Errorf("connector cannot be nil")
	}
	if resolveCredentials == nil {
		errors["retrieveCredentials"] = fmt.Errorf("retrieveCredentials cannot be nil")
	}
	if listener == nil {
		errors["listener"] = fmt.Errorf("listener cannot be nil")
	}

	if err := errors.Filter(); err != nil {
		return nil, err
	}

	return &proxyService{
		connector:           connector,
		retrieveCredentials: resolveCredentials,
		listener:            listener,
		done:                false,
		error:               make(chan error),
	}, nil
}

func (proxy *proxyService) handleConnection(clientConn net.Conn) error {
	var targetConn net.Conn
	defer func() {
		if clientConn != nil {
			_ = clientConn.Close()
		}
		if clientConn != nil {
			_ = targetConn.Close()
		}
	}()

	backendCredentials, err := proxy.retrieveCredentials()
	defer zeroizeCredentials(backendCredentials)
	if err != nil {
		return err
	}

	targetConn, err = proxy.connector(clientConn, backendCredentials)
	if err != nil {
		return err
	}

	return <-duplexStream(clientConn, targetConn)
}

func (proxy *proxyService) OperationalError() chan error {
	return proxy.error
}

// Start initiates the net.Listener to listen for incoming connections
func (proxy *proxyService) Start() error {
	go func() { // n go routines for n tcp proxyservices
		// TODO: add mutex to done
		for !proxy.done {
			// TODO: can accepts happen in parallel ?
			conn, err := proxy.listener.Accept()
			if err != nil {
				proxy.error <- err
				return
			}
			go func() {
				err := proxy.handleConnection(conn)
				proxy.error <- err
			}()
		}
	}()

	return nil
}

// Stop terminates proxyService by closing the listening net.Listener
func (proxy *proxyService) Stop() error {
	proxy.done = true
	return proxy.listener.Close()
}
