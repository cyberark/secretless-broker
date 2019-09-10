package ssh

import (
	"io"
	"log"
	"reflect"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
)

// ServerConfig is the configuration info for the target server
type ServerConfig struct {
	Network      string
	Address      string
	ClientConfig ssh.ClientConfig
}

// Handler contains the configuration and channels
type Handler struct {
	plugin_v1.BaseHandler
	Channels <-chan ssh.NewChannel
}

func (h *Handler) serverConfig() (config ServerConfig, err error) {
	var values map[string][]byte

	// TODO: Ensure that we don't print credentials here before uncommenting
	// Issue: https://github.com/cyberark/secretless-broker/issues/593
	//
	// log.Printf("%s", h.GetConfig().Credentials)

	if values, err = h.Resolver.Resolve(h.GetConfig().Credentials); err != nil {
		return
	}

	if h.GetConfig().Debug {
		keys := reflect.ValueOf(values).MapKeys()
		log.Printf("SSH backend connection parameters: %s", keys)
	}

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

	if h.HandlerConfig.Debug {
		log.Printf("Trying to connect with user: %s", config.ClientConfig.User)
	}

	if hostKeyStr, ok := values["hostKey"]; ok {
		var hostKey ssh.PublicKey
		if hostKey, err = ssh.ParsePublicKey([]byte(hostKeyStr)); err != nil {
			log.Printf("Unable to parse public key: %v", err)
			return
		}
		config.ClientConfig.HostKeyCallback = ssh.FixedHostKey(hostKey)
	} else {
		log.Printf("WARN: No SSH hostKey specified. Secretless will accept any backend host key!")
		config.ClientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	if privateKeyStr, ok := values["privateKey"]; ok {
		var signer ssh.Signer
		if signer, err = ssh.ParsePrivateKey([]byte(privateKeyStr)); err != nil {
			log.Printf("Unable to parse private key: %v", err)
			return
		}
		config.ClientConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}

	return
}

// Run opens the connection to the target server and proxies requests
func (h *Handler) Run() {
	var err error
	var serverConfig ServerConfig
	var server ssh.Conn

	if serverConfig, err = h.serverConfig(); err != nil {
		log.Fatalf("ERROR: Could not resolve server config: %s\n", err)
	}

	// TODO: Ensure that we don't print credentials here before uncommenting
	// Issue: https://github.com/cyberark/secretless-broker/issues/593
	//
	// if h.HandlerConfig.Debug {
	// 	log.Printf("Using config\n%v", serverConfig.ClientConfig)
	// }

	if server, err = ssh.Dial(serverConfig.Network, serverConfig.Address, &serverConfig.ClientConfig); err != nil {
		log.Printf("Failed to dial SSH backend '%s': %s", serverConfig.Address, err)
		return
	}

	// Service the incoming Channel channel.
	for newChannel := range h.Channels {
		serverChannel, serverRequests, err := server.OpenChannel(newChannel.ChannelType(), newChannel.ExtraData())
		if err != nil {
			sshError := err.(*ssh.OpenChannelError)
			if err := newChannel.Reject(sshError.Reason, sshError.Message); err != nil {
				log.Printf("Failed to send new channel rejection : %s", err)
			}
			return
		}

		clientChannel, clientRequests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Failed to accept client channel : %s", err)
			serverChannel.Close()
			return
		}

		go func() {
			for clientRequest := range clientRequests {
				log.Printf("Client request : %s", clientRequest.Type)
				ok, err := serverChannel.SendRequest(clientRequest.Type, clientRequest.WantReply, clientRequest.Payload)
				if err != nil {
					log.Printf("WARN: Failed to send client request to server channel : %s", err)
				}
				if clientRequest.WantReply {
					log.Printf("Server reply is %v", ok)
				}
			}
		}()

		go func() {
			for serverRequest := range serverRequests {
				log.Printf("Server request : %s", serverRequest.Type)
				ok, err := clientChannel.SendRequest(serverRequest.Type, serverRequest.WantReply, serverRequest.Payload)
				if err != nil {
					log.Printf("WARN: Failed to send server request to client channel : %s", err)
				}
				if serverRequest.WantReply {
					log.Printf("Client reply is %v", ok)
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
					log.Printf("Client channel is closed")
					time.Sleep(softDelay)
					serverChannel.Close()
					return
				}
				_, err = serverChannel.Write(data[0:len])
				if err != nil {
					log.Printf("Error writing %d bytes to server channel : %s", len, err)
				}
			}
		}()

		go func() {
			for {
				data := make([]byte, 1024)
				len, err := serverChannel.Read(data)
				if err == io.EOF {
					log.Printf("Server channel is closed")
					time.Sleep(softDelay)
					clientChannel.Close()
					return
				}
				_, err = clientChannel.Write(data[0:len])
				if err != nil {
					log.Printf("Error writing %d bytes to client channel : %s", len, err)
				}
			}
		}()
	}
}

// HandlerFactory instantiates a handler given HandlerOptions
func HandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &Handler{
		BaseHandler: plugin_v1.NewBaseHandler(options),
		Channels:    options.Channels,
	}

	handler.Run()

	return handler
}
