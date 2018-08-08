package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"

	"github.com/conjurinc/secretless-broker/internal/pkg/util"
	"github.com/conjurinc/secretless-broker/pkg/secretless/config"
	plugin_v1 "github.com/conjurinc/secretless-broker/pkg/secretless/plugin/v1"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Listener accepts SSH connections and MITMs them using a Handler.
//
// NOTE: This MITM approach to SSH is experimental. The ssh-agent approach is
// better validated and probably better all-around.
type Listener struct {
	Config         config.Listener
	EventNotifier  plugin_v1.EventNotifier
	HandlerConfigs []config.Handler
	NetListener    net.Listener
	Resolver       plugin_v1.Resolver
	RunHandlerFunc func(id string, options plugin_v1.HandlerOptions) plugin_v1.Handler
}

// HandlerHasCredentials validates that a handler has all necessary credentials.
type handlerHasCredentials struct {
}

// Validate checks that a handler has all necessary credentials.
func (hhc handlerHasCredentials) Validate(value interface{}) error {
	hs := value.([]config.Handler)
	errors := validation.Errors{}
	for i, h := range hs {
		for _, credential := range [...]string{"address", "privateKey"} {
			if !h.HasCredential(credential) {
				errors[strconv.Itoa(i)] = fmt.Errorf(credential)
			}
		}
	}
	return errors.Filter()
}

// Validate verifies the completeness and correctness of the Listener.
func (l Listener) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.HandlerConfigs, validation.Required),
		validation.Field(&l.HandlerConfigs, handlerHasCredentials{}),
	)
}

// _GenerateSSHKeys generates a new private and public keypair
func _GenerateSSHKeys(keyPath string) error {
	// Create new private key of length 2048
	// TODO: Add capability to specify different sizes
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Generate a PEM structure using the private key
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Create the destination private key file
	privateKeyFile, err := os.Create(keyPath)
	defer privateKeyFile.Close()
	if err != nil {
		return err
	}

	// Write out the PEM object to the private key file
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return err
	}

	// Get our public key part from the private key we generated
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	log.Printf("New host key fingerprint: %s", ssh.FingerprintSHA256(publicKey))

	// Write the public key into the provided path
	publicKeyPath := keyPath + ".pub"
	return ioutil.WriteFile(publicKeyPath,
		ssh.MarshalAuthorizedKey(publicKey),
		0644)
}

// Listen accepts SSH connections and MITMs them using a Handler.
func (l *Listener) Listen() {
	expectedHostKeyPath := "/tmp/id_rsa"

	// Generate a host key if one isn't present in ./tmp/id_rsa
	// TODO: Be able to use secretless.yml-provided host key
	if _, err := os.Stat(expectedHostKeyPath); err != nil {
		log.Printf("Could not find pre-existing host key at %s. Generating...", expectedHostKeyPath)
		if err := _GenerateSSHKeys(expectedHostKeyPath); err != nil {
			log.Fatal("Failed to create host key: ", err)
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
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	serverConfig.AddHostKey(private)

	for {
		nConn, err := util.Accept(l)
		if err != nil {
			log.Printf("Failed to accept incoming connection: ", err)
			continue
		}

		// https://godoc.org/golang.org/x/crypto/ssh#NewServerConn
		conn, chans, reqs, err := ssh.NewServerConn(nConn, serverConfig)
		if err != nil {
			log.Printf("Failed to handshake: %s", err)
			continue
		}
		log.Printf("New connection accepted for user %s from %s", conn.User(), conn.RemoteAddr())

		// The incoming Request channel must be serviced.
		go func() {
			for req := range reqs {
				log.Printf("Global SSH request : %s", req)
			}
		}()

		// Serve the first Handler which is attached to this listener
		if len(l.HandlerConfigs) == 0 {
			log.Panicf("No ssh handler is available")
		}

		handlerOptions := plugin_v1.HandlerOptions{
			HandlerConfig:    l.HandlerConfigs[0],
			Channels:         chans,
			ClientConnection: nil,
			EventNotifier:    l.EventNotifier,
			Resolver:         l.Resolver,
		}

		// TODO: Kill connection to client when backend fails to be contacted
		l.RunHandlerFunc("ssh", handlerOptions)
	}
}

// GetConfig implements plugin_v1.Listener
func (l *Listener) GetConfig() config.Listener {
	return l.Config
}

// GetListener implements plugin_v1.Listener
func (l *Listener) GetListener() net.Listener {
	return l.NetListener
}

// GetHandlers implements plugin_v1.Listener
func (l *Listener) GetHandlers() []plugin_v1.Handler {
	return nil
}

// GetConnections implements plugin_v1.Listener
func (l *Listener) GetConnections() []net.Conn {
	return nil
}

// GetNotifier implements plugin_v1.Listener
func (l *Listener) GetNotifier() plugin_v1.EventNotifier {
	return l.EventNotifier
}

// GetName implements plugin_v1.Listener
func (l *Listener) GetName() string {
	return "ssh"
}

// Shutdown implements plugin_v1.Listener
func (l *Listener) Shutdown() error {
	// TODO: Clean up all handlers
	return l.NetListener.Close()
}

// ListenerFactory returns a Listener created from options
func ListenerFactory(options plugin_v1.ListenerOptions) plugin_v1.Listener {
	return &Listener{
		Config:         options.ListenerConfig,
		EventNotifier:  options.EventNotifier,
		HandlerConfigs: options.HandlerConfigs,
		NetListener:    options.NetListener,
		Resolver:       options.Resolver,
		RunHandlerFunc: options.RunHandlerFunc,
	}
}
