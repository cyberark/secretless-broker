package proxy

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/kgilpin/secretless-pg/config"
	"github.com/kgilpin/secretless-pg/connect"
	"github.com/kgilpin/secretless-pg/protocol"
)

type Proxy struct {
	Config config.Config
}

type ClientOptions struct {
	User     string
	Database string
	Options  map[string]string
}

type Handler struct {
	Config            config.Config
	Client            net.Conn
	Backend           net.Conn
	ClientOptions     ClientOptions
	BackendConnection BackendConnection
	BackendConfig     config.BackendConfig
}

func stream(source, dest net.Conn) {
	for {
		if message, length, err := connect.Receive(source); err == nil {
			_, err = connect.Send(dest, message[:length])
			if err != nil {
				log.Printf("Error sending to %s : %s", dest.RemoteAddr(), err)
				if err == io.EOF {
					return
				}
			}
		} else {
			if err == io.EOF {
				log.Printf("Connection closed from %s", source.RemoteAddr())
				return
			} else {
				log.Printf("Error reading from %s : %s", source.RemoteAddr(), err)
			}
		}
	}
}

func (self *Handler) Pipe() {
	log.Printf("Connecting client %s to backend %s", self.Client.RemoteAddr(), self.Backend.RemoteAddr())

	go stream(self.Client, self.Backend)
	go stream(self.Backend, self.Client)
}

func (self *Handler) ConnectToBackend() error {
	var connection net.Conn
	var err error

	if connection, err = connect.Connect(self.BackendConfig.Address); err != nil {
		return err
	}

	log.Print("Sending startup message")
	startupMessage := protocol.CreateStartupMessage(self.BackendConfig.Username, self.ClientOptions.Database, self.BackendConfig.Options)

	connection.Write(startupMessage)

	response := make([]byte, 4096)
	connection.Read(response)

	log.Print("Authenticating")
	message, authenticated := connect.HandleAuthenticationRequest(self.BackendConfig.Username, self.BackendConfig.Password, connection, response)

	if !authenticated {
		return fmt.Errorf("Authentication failed")
	}

	log.Printf("Successfully connected to '%s'", self.BackendConfig.Address)

	if _, err = connect.Send(self.Client, message); err != nil {
		return err
	}

	self.Backend = connection

	return nil
}

func (self *Handler) Abort(err error) {
	pgError := protocol.Error{
		Severity: protocol.ErrorSeverityFatal,
		Code:     protocol.ErrorCodeInternalError,
		Message:  err.Error(),
	}
	connect.Send(self.Client, pgError.GetMessage())
	return
}

func (self *Handler) Run() {
	var authenticationError, err error
	var abort bool
	var backendConfig *config.BackendConfig

	if err = self.Startup(); err != nil {
		self.Abort(err)
		return
	}

	abort, authenticationError, err = self.Authenticate()

	if authenticationError != nil {
		pgError := protocol.Error{
			Severity: protocol.ErrorSeverityFatal,
			Code:     protocol.ErrorCodeInvalidPassword,
			Message:  authenticationError.Error(),
		}
		connect.Send(self.Client, pgError.GetMessage())
		return
	}

	if err != nil {
		self.Abort(err)
		return
	}

	/* Benign abort condidition in authentication */
	if abort {
		return
	}

	if backendConfig, err = self.BackendConnection.Configure(); err != nil {
		self.Abort(err)
		return
	}
	self.BackendConfig = *backendConfig

	if err = self.ConnectToBackend(); err != nil {
		self.Abort(err)
		return
	}

	self.Pipe()
}

func (self *Proxy) Run() {
	proxyListener, err := net.Listen("tcp", self.Config.Address)
	if err == nil {
		log.Printf("Server listening on: %s", proxyListener.Addr())
		for {
			var client net.Conn

			if client, err = proxyListener.Accept(); err != nil {
				log.Println(err)
				continue
			}

			handler := Handler{Config: self.Config, Client: client}
			go handler.Run()
		}
	} else {
		log.Fatal(err)
	}
}
