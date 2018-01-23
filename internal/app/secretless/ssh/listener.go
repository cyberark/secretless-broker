package ssh

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"

	"github.com/conjurinc/secretless/internal/pkg/provider"
	"github.com/conjurinc/secretless/pkg/secretless/config"
)

type Listener struct {
	Config    config.Listener
	Handlers  []config.Handler
	Providers []provider.Provider
	Listener  net.Listener
}

func (self *Listener) Listen() {
	serverConfig := &ssh.ServerConfig{
    PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	    return nil, nil
    },
    PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
    	return nil, fmt.Errorf("Public key authentication is not supported")
    },
	}

	privateBytes, err := ioutil.ReadFile("./tmp/id_rsa")
	if err != nil {
	    log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
	    log.Fatal("Failed to parse private key: ", err)
	}

	serverConfig.AddHostKey(private)

	for {
		nConn, err := self.Listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: ", err)
			return
		}

		// https://godoc.org/golang.org/x/crypto/ssh#NewServerConn
		conn, chans, reqs, err := ssh.NewServerConn(nConn, serverConfig)
		if err != nil {
		    log.Printf("Failed to handshake: %s", err)
		    return
		}
		log.Printf("New connection accepted for user %s from %s", conn.User(), conn.RemoteAddr())

		// The incoming Request channel must be serviced.
		go func() {
			  for req := range reqs {
			  	log.Printf("Global SSH request : %s", req)
				}		
			}()
		
		// Serve the first Handler which is attached to this listener
		var selectedHandler *config.Handler
		for _, handler := range self.Handlers {
			listener := handler.Listener
			if listener == "" {
				listener = handler.Name
			}

			if listener == self.Config.Name {
				selectedHandler = &handler
				break
			}
		}

		if selectedHandler != nil {
			handler := &Handler{Providers: self.Providers, Config: *selectedHandler, Channels: chans}
			handler.Run()
		} else {
			log.Printf("No ssh handler is available for this connection!")
		}
	}
}
