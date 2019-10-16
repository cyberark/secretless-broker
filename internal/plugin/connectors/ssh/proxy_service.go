package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

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

			tcpConn := nConn.(*net.TCPConn)
			logger.Debugf("SSH Client connected. ClientIP=%v", tcpConn.RemoteAddr())


			go func() {
				backendCredentials, err := proxy.retrieveCredentials()
				defer internal.ZeroizeCredentials(backendCredentials)
				if err != nil {
					logger.Errorf("Failed on retrieve credentials: %s", err)
					return
				}

				clientConfig := &ssh.ClientConfig{}
				var address string
				if addressBytes, ok := backendCredentials["address"]; ok {
					address = string(addressBytes)
					if !strings.Contains(address, ":") {
						address = address + ":22"
					}
				}

				if user, ok := backendCredentials["user"]; ok {
					clientConfig.User = string(user)
				}

				logger.Debugf("Trying to connect with user: %s", clientConfig.User)

				if hostKeyStr, ok := backendCredentials["hostKey"]; ok {
					var hostKey ssh.PublicKey
					if hostKey, err = ssh.ParsePublicKey([]byte(hostKeyStr)); err != nil {
						logger.Errorf("Unable to parse public key: %v", err)
						return
					}
					clientConfig.HostKeyCallback = ssh.FixedHostKey(hostKey)
				} else {
					clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
				}

				if password, ok := backendCredentials["password"]; ok {
					clientConfig.Auth = append(clientConfig.Auth, ssh.Password(string(password)))
				}

				if privateKeyBytes, ok := backendCredentials["privateKey"]; ok {
					var signer ssh.Signer
					if signer, err = ssh.ParsePrivateKey([]byte(privateKeyBytes)); err != nil {
						logger.Debugf("Unable to parse private key: %v", err)
						return
					}

					clientConfig.Auth = append(clientConfig.Auth, ssh.PublicKeys(signer))
				}

				p, err := newSSHProxyConn(
					tcpConn,
					serverConfig,
					clientConfig,
					address,
					)
				if err != nil {
					logger.Errorln("Connection from %v closed. %v", tcpConn.RemoteAddr(), err)
					return
				}
				logger.Infof("Establish a proxy connection between %v and %v", tcpConn.RemoteAddr(), p.DestinationHost)

				err = p.Wait()
				logger.Debugf("Connection from %v closed. %v", tcpConn.RemoteAddr(), err)
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

func newSSHProxyConn(
	conn net.Conn,
	serverConfig *ssh.ServerConfig,
	clientConfig *ssh.ClientConfig,
	upstreamHostAndPort string,
) (proxyConn *ssh.ProxyConn, err error) {
	d, err := ssh.NewDownstreamConn(conn, serverConfig)
	if err != nil {
		return nil, err
	}
	defer func() {
		if proxyConn == nil {
			d.Close()
		}
	}()

	authRequestMsg, err := d.GetAuthRequestMsg()
	if err != nil {
		return nil, err
	}

	// use client provided user if client config is empty
	if clientConfig.User == "" {
		clientConfig.User = authRequestMsg.User
	}

	upConn, err := net.Dial("tcp", upstreamHostAndPort)
	if err != nil {
		return nil, err
	}

	u, err := ssh.NewUpstreamConn(upConn, clientConfig)
	if err != nil {
		return nil, err
	}
	defer func() {
		if proxyConn == nil {
			u.Close()
		}
	}()
	p := &ssh.ProxyConn{
		Upstream:   u,
		Downstream: d,
	}

	if err = p.AuthenticateProxyConn(clientConfig); err != nil {
		return nil, err
	}

	return p, nil
}
