package ssh

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/ssh"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// ServerConfig is the configuration info for the target server
type ServerConfig struct {
	Network      string
	Address      string
	ClientConfig ssh.ClientConfig
}

// ServiceConnector contains the configuration and channels
type ServiceConnector struct {
	channels <-chan ssh.NewChannel
	logger   log.Logger
}

func (h *ServiceConnector) serverConfig(values map[string][]byte) (config ServerConfig, err error) {
	keys := reflect.ValueOf(values).MapKeys()
	h.logger.Debugf("SSH backend connection parameters: %s", keys)

	config.Network = "tcp"
	if address, ok := values["address"]; ok {
		config.Address = string(address)
		if !strings.Contains(config.Address, ":") {
			config.Address = config.Address + ":22"
		}
	}

	// XXX: Should this be the user that the client was trying to connect as?
	config.ClientConfig.User = "root"
	if user, ok := values["user"]; ok {
		config.ClientConfig.User = string(user)

	}

	h.logger.Debugf("Trying to connect with user: %s", config.ClientConfig.User)

	if hostKeyStr, ok := values["hostKey"]; ok {
		var hostKey ssh.PublicKey
		if hostKey, err = ssh.ParsePublicKey([]byte(hostKeyStr)); err != nil {
			h.logger.Debugf("Unable to parse public key: %v", err)
			return
		}
		config.ClientConfig.HostKeyCallback = ssh.FixedHostKey(hostKey)
	} else {
		h.logger.Warnf("No SSH hostKey specified. Secretless will accept any backend host key!")
		config.ClientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	if privateKeyStr, ok := values["privateKey"]; ok {
		var signer ssh.Signer
		if signer, err = ssh.ParsePrivateKey([]byte(privateKeyStr)); err != nil {
			h.logger.Debugf("Unable to parse private key: %v", err)
			return
		}
		config.ClientConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}

	return
}

// Connect opens the connection to the target server and proxies requests
func (h *ServiceConnector) Connect(
	credentialValuesByID connector.CredentialValuesByID,
) error {
	var err error
	var serverConfig ServerConfig
	var server ssh.Conn

	errors := validation.Errors{}
	for _, credential := range [...]string{"address", "privateKey"} {
		if _, hasCredential := credentialValuesByID[credential]; !hasCredential {
			errors[credential] = fmt.Errorf("must have credential '%s'", credential)
		}
	}

	if err := errors.Filter(); err != nil {
		return err
	}

	if serverConfig, err = h.serverConfig(credentialValuesByID); err != nil {
		return fmt.Errorf("could not resolve server config: '%s'", err)
	}

	if server, err = ssh.Dial(serverConfig.Network, serverConfig.Address, &serverConfig.ClientConfig); err != nil {
		return fmt.Errorf("failed to dial SSH backend '%s': %s", serverConfig.Address, err)
	}

	// Service the incoming Channel channel.
	for newChannel := range h.channels {
		serverChannel, serverRequests, err := server.OpenChannel(newChannel.ChannelType(), newChannel.ExtraData())
		if err != nil {
			sshError := err.(*ssh.OpenChannelError)
			if err := newChannel.Reject(sshError.Reason, sshError.Message); err != nil {
				h.logger.Errorf("Failed to send new channel rejection : %s", err)
			}
			return err
		}

		clientChannel, clientRequests, err := newChannel.Accept()
		if err != nil {
			h.logger.Errorf("Failed to accept client channel : %s", err)
			serverChannel.Close()
			return err
		}

		go func() {
			for clientRequest := range clientRequests {
				h.logger.Debugf("Client request : %s", clientRequest.Type)
				ok, err := serverChannel.SendRequest(clientRequest.Type, clientRequest.WantReply, clientRequest.Payload)
				if err != nil {
					h.logger.Warnf("Failed to send client request to server channel : %s", err)
				}
				if clientRequest.WantReply {
					h.logger.Debugf("Server reply is %v", ok)
				}
			}
		}()

		go func() {
			for serverRequest := range serverRequests {
				h.logger.Debugf("Server request : %s", serverRequest.Type)
				ok, err := clientChannel.SendRequest(serverRequest.Type, serverRequest.WantReply, serverRequest.Payload)
				if err != nil {
					h.logger.Debugf("WARN: Failed to send server request to client channel : %s", err)
				}
				if serverRequest.WantReply {
					h.logger.Debugf("Client reply is %v", ok)
				}
			}
		}()

		// This delay is to prevent closing of channels on the other side
		// too early when we receive an EOF but have not had the chance to
		// pass that on to the client/server.
		// TODO: Maybe use a better logic for handling EOF conditions
		softDelay := time.Second * 2

		go func() {
			for {
				data := make([]byte, 1024)
				len, err := clientChannel.Read(data)
				if err == io.EOF {
					h.logger.Debugf("Client channel is closed")
					time.Sleep(softDelay)
					serverChannel.Close()
					return
				}
				_, err = serverChannel.Write(data[0:len])
				if err != nil {
					h.logger.Debugf("Error writing %d bytes to server channel : %s", len, err)
				}
			}
		}()

		go func() {
			for {
				data := make([]byte, 1024)
				len, err := serverChannel.Read(data)
				if err == io.EOF {
					h.logger.Debugf("Server channel is closed")
					time.Sleep(softDelay)
					clientChannel.Close()
					return
				}
				_, err = clientChannel.Write(data[0:len])
				if err != nil {
					h.logger.Debugf("Error writing %d bytes to client channel : %s", len, err)
				}
			}
		}()
	}

	return nil
}
