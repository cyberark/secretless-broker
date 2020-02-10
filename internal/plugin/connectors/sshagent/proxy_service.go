package sshagent

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/ssh/agent"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/pkg/secretless/log"
)

type proxyService struct {
	done                bool
	listener            net.Listener
	logger              log.Logger
	keyring             agent.Agent
	retrieveCredentials internal.CredentialsRetriever
}

// NewProxyService constructs a new instance of a SSH Agent ProxyService. The
// constructor takes a Listener, Logger and CredentialResolver.
// An SSH Agent ProxyService serves ssh-agent protocol requests from
// an in-memory keyring.
//
// Typical usage has the SSH client delegating auth to the agent e.g.:
// SSH_AUTH_SOCK=/sock/.agent ssh -T git@github.com
//
// NOTE: The keyring is populated at proxy service startup and so is unable to cope
// with rotation
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
		keyring:             agent.NewKeyring(),
	}, nil
}

func newPrivateKey(pemStr []byte) (interface{}, error) {
	block, _ := pem.Decode(pemStr)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported key type %q", block.Type)
	}
}

func (proxy *proxyService) populateKeyring(
	backendCredentials map[string][]byte,
) error {
	var err error

	key := agent.AddedKey{}

	// select between rsa and ecdsa
	if rsa, ok := backendCredentials["rsa"]; ok {
		key.PrivateKey, err = newPrivateKey(rsa)
	} else if ecdsa, ok := backendCredentials["ecdsa"]; ok {
		key.PrivateKey, err = newPrivateKey(ecdsa)
	} else {
		err = fmt.Errorf("neither 'rsa' nor 'ecdsa' credentials found")
	}

	if err != nil {
		return err
	}

	// TODO: neither comment, lifetime, nor confirm is a credential.
	// Maybe Handler needs a mechanism for these types of non-secret configuration options.
	if comment, ok := backendCredentials["comment"]; ok {
		key.Comment = string(comment)
	}
	if lifetime, ok := backendCredentials["lifetime"]; ok {
		var lt uint64
		lt, err = strconv.ParseUint(string(lifetime), 10, 32)
		if err != nil {
			return err
		}
		key.LifetimeSecs = uint32(lt)
	}
	if confirm, ok := backendCredentials["confirm"]; ok {
		key.ConfirmBeforeUse, err = strconv.ParseBool(string(confirm))
		if err != nil {
			return err
		}
	}

	return proxy.keyring.Add(key)
}

// Start initiates the net.Listener to listen for incoming connections
func (proxy *proxyService) Start() error {
	logger := proxy.logger

	logger.Infof("Starting service")

	if proxy.done {
		return fmt.Errorf("cannot call Start on stopped ProxyService")
	}

	// TODO: this proxy service only fetches credentials at the outset
	//   so can not cope with rotation
	backendCredentials, err := proxy.retrieveCredentials()
	defer internal.ZeroizeCredentials(backendCredentials)
	if err != nil {
		return fmt.Errorf("failed on retrieve credentials: %s", err)
	}

	err = proxy.populateKeyring(backendCredentials)
	if err != nil {
		return err
	}

	go func() {
		for !proxy.done {
			conn, err := proxy.listener.Accept()
			if err != nil {
				logger.Errorf("Failed on accept connection: %s", err)
				return
			}

			go func() {
				proxy.logger.Debugf("Serving SSH Agent connection on %v", conn.LocalAddr())
				err := agent.ServeAgent(proxy.keyring, conn)
				if err != nil && err != io.EOF {
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
	// TODO: we should verify with a channel that outer goroutine actually stops.
	return proxy.listener.Close()
}
