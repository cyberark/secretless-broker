package tcp

import (
	"fmt"
	"io"
	"net"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

const closedConnectionErrString = "use of closed network connection"

func duplexStream(
	source io.ReadWriter,
	destination io.ReadWriter,
) (sourceErrorChan <-chan error, destinationErrorChan <-chan error) {
	_sourceErrorChan := make(chan error)
	_destinationErrorChan := make(chan error)

	go func() {
		_sourceErrorChan <- stream(source, destination)
	}()
	go func() {
		_destinationErrorChan <- stream(destination, source)
	}()

	sourceErrorChan = _sourceErrorChan
	destinationErrorChan = _destinationErrorChan
	return
}

func stream(source io.Reader, destination io.Writer) error {
	_, err := io.Copy(destination, source)
	return err
}

type proxyService struct {
	connector           tcp.Connector
	done                bool
	listener            net.Listener
	logger              log.Logger
	retrieveCredentials internal.CredentialsRetriever
}

// NewProxyService constructs a new instance of a TCP ProxyService. The
// constructor takes a TCP Connector, CredentialResolver and Listener.
// A TCP ProxyService is able to Connect with Credentials then subsequently stream
// bytes between client and target service
func NewProxyService(
	connector tcp.Connector,
	listener net.Listener,
	logger log.Logger,
	retrieveCredentials internal.CredentialsRetriever,
) (internal.Service, error) {
	errors := validation.Errors{}

	if connector == nil {
		errors["connector"] = fmt.Errorf("connector cannot be nil")
	}
	if retrieveCredentials == nil {
		errors["retrieveCredentials"] = fmt.Errorf("retrieveCredentials cannot be nil")
	}
	if listener == nil {
		errors["logger"] = fmt.Errorf("logger cannot be nil")
	}
	if logger == nil {
		errors["listener"] = fmt.Errorf("listener cannot be nil")
	}

	if err := errors.Filter(); err != nil {
		return nil, err
	}

	return &proxyService{
		connector:           connector,
		retrieveCredentials: retrieveCredentials,
		listener:            listener,
		logger:              logger,
		done:                false,
	}, nil
}

func closeConn(conn net.Conn, connDescription string, logger log.Logger) {
	if conn == nil {
		return
	}
	err := conn.Close()
	if err != nil {
		logger.Warnf("Failed on closing %s connection: %s", connDescription, err)
	}
}

func (proxy *proxyService) handleConnection(clientConn net.Conn) error {
	var targetConn net.Conn
	logger := proxy.logger

	defer func() {
		closeConn(clientConn, "client", logger)
		closeConn(targetConn, "target", logger)
	}()

	backendCredentials, err := proxy.retrieveCredentials()
	// zeroize credentials if we exit early due to an error
	defer internal.ZeroizeCredentials(backendCredentials)
	if err != nil {
		return errors.Wrap(err, "failed on retrieve credentials")
	}

	logger.Debugf("New connection on %v.\n", clientConn.LocalAddr())

	targetConn, err = proxy.connector.Connect(clientConn, backendCredentials)
	if err != nil {
		return errors.Wrap(err, "failed on connect")
	}

	// immediately zeroize credentials after connecting
	internal.ZeroizeCredentials(backendCredentials)

	logger.Debugf("Proxying connection on %v to %v.\n", clientConn.LocalAddr(), targetConn.RemoteAddr())

	clientErrChan, destErrChan := duplexStream(clientConn, targetConn)

	var closer string
	select {
	case err = <-clientErrChan:
		closer = "client"
	case err = <-destErrChan:
		closer = "target"
	}

	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf(
				`connection on %v failed while streaming from %s connection`,
				clientConn.LocalAddr(),
				closer,
			),
		)
	}

	logger.Debugf("Connection on %v closed by %s.\n", clientConn.LocalAddr(), closer)
	return nil
}

// Start initiates the net.Listener to listen for incoming connections
func (proxy *proxyService) Start() error {
	logger := proxy.logger

	logger.Infof("Starting service")

	if proxy.done {
		return fmt.Errorf("cannot call Start on stopped ProxyService")
	}

	go func() { // n go routines for n tcp ProxyServices
		for !proxy.done {
			// TODO: can accepts happen in parallel ?
			conn, err := proxy.listener.Accept()

			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == closedConnectionErrString {
				logger.Info("Listener closed")
				return
			}
			if err != nil {
				logger.Errorf("Failed on accept connection: %s", err)
				return
			}
			go func() {
				err := proxy.handleConnection(conn)

				if err == nil {
					return
				}

				// io.EOF means connection was closed
				if errors.Cause(err) == io.EOF {
					err = errors.Wrap(
						err,
						fmt.Sprintf(
							"connection closed early on %v\n",
							conn.LocalAddr(),
						),
					)
				}

				logger.Errorf("Failed on handle connection: %s", err)
			}()
		}
	}()

	return nil
}

// Stop terminates proxyService by closing the listening net.Listener
func (proxy *proxyService) Stop() error {
	proxy.logger.Infof("Stopping service")
	proxy.done = true

	return proxy.listener.Close()
}
