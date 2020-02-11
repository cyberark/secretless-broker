package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	validation "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/ssh"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

// An SSH ProxyService accepts SSH connections and MITMs them.
//
// NOTE: This MITM approach to SSH is experimental. The ssh-agent approach is
// better validated and probably better all-around.
type proxyService struct {
	connector           tcp.Connector
	done                bool
	listener            net.Listener
	logger              log.Logger
	retrieveCredentials internal.CredentialsRetriever
}

// NewProxyService constructs a new instance of a SSH ProxyService. The
// constructor takes a Listener, Logger and CredentialResolver.
// A SSH ProxyService is able to Connect with Credentials then subsequently stream
// bytes between client and target service
func NewProxyService(
	listener net.Listener,
	logger log.Logger,
	retrieveCredentials internal.CredentialsRetriever,
) (internal.Service, error) {
	errors := validation.Errors{}

	if retrieveCredentials == nil {
		errors["retrieveCredentials"] = fmt.Errorf("retrieveCredentials cannot be nil")
	}
	if logger == nil {
		errors["logger"] = fmt.Errorf("logger cannot be nil")
	}
	if listener == nil {
		errors["listener"] = fmt.Errorf("listener cannot be nil")
	}

	if err := errors.Filter(); err != nil {
		return nil, err
	}

	return &proxyService{
		retrieveCredentials: retrieveCredentials,
		listener:            listener,
		logger:              logger,
		done:                false,
	}, nil
}

func (proxy *proxyService) handleConnections(channels <-chan ssh.NewChannel) error {
	var connector = ServiceConnector{
		channels: channels,
		logger:   proxy.logger,
	}

	backendCredentials, err := proxy.retrieveCredentials()
	defer internal.ZeroizeCredentials(backendCredentials)
	if err != nil {
		return fmt.Errorf("failed on retrieve credentials: %s", err)
	}

	return connector.Connect(backendCredentials)
}

// Start initiates the net.Listener to listen for incoming connections
// Listen accepts SSH connections and MITMs them using a ServiceConnector.
func (proxy *proxyService) Start() error {
	logger := proxy.logger

	logger.Infof("Starting service")

	if proxy.done {
		return fmt.Errorf("cannot call Start on stopped ProxyService")
	}

	go func() {
		expectedHostKeyPath := "/tmp/id_rsa"

		// Generate a host key if one isn't present in ./tmp/id_rsa
		// TODO: Be able to use secretless.yml-provided host key
		if _, err := os.Stat(expectedHostKeyPath); err != nil {
			logger.Debugf("Could not find pre-existing host key at %s. Generating...", expectedHostKeyPath)
			if err := _GenerateSSHKeys(expectedHostKeyPath); err != nil {
				logger.Panicf("Failed to create host key: ", err)
			}
		}

		serverConfig := &ssh.ServerConfig{
			NoClientAuth: true,
			PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
				return nil, nil
			},
			PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
				return nil, nil
			},
		}

		privateBytes, err := ioutil.ReadFile(expectedHostKeyPath)
		if err != nil {
			logger.Panicf("Failed to load private key: ", err)
		}

		private, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			logger.Panicf("Failed to parse private key: ", err)
		}

		serverConfig.AddHostKey(private)

		// TODO: is it possible to use the duplex func to stream ?
		for !proxy.done {
			nConn, err := proxy.listener.Accept()
			if err != nil {
				logger.Debugf("Failed on accept connection: %s", err)
				return
			}

			// https://godoc.org/golang.org/x/crypto/ssh#NewServerConn
			conn, chans, reqs, err := ssh.NewServerConn(nConn, serverConfig)
			if err != nil {
				logger.Debugf("Failed to handshake: %s", err)
				continue
			}
			logger.Debugf(
				"New connection accepted for user %s from %s",
				conn.User(),
				conn.RemoteAddr(),
			)

			// The incoming Request channel must be serviced.
			go func() {
				for req := range reqs {
					logger.Debugf("Global SSH request : %v", req)
				}
			}()

			go func() {
				if err := proxy.handleConnections(chans); err != nil {
					logger.Errorf("Failed on handle connection: %s", err)
					return
				}

				logger.Debugf("Connection closed on %v", conn.LocalAddr())
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
