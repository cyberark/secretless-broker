package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"os"
	"strings"
	"encoding/base64"

	"github.com/kgilpin/secretless-pg/config"
	"github.com/kgilpin/secretless-pg/conjur"
	"github.com/kgilpin/secretless-pg/connect"
	"github.com/kgilpin/secretless-pg/protocol"
)

var HostUsername = os.Getenv("CONJUR_AUTHN_LOGIN")
var HostAPIKey = os.Getenv("CONJUR_AUTHN_API_KEY")

func initialize(client net.Conn) (map[string]string, error) {
	log.Printf("Handling connection %v", client)

	/* Get the client startup message. */
	message, length, err := connect.Receive(client)
	if err != nil {
		return nil, fmt.Errorf("Error receiving startup message from client: %s", err)
	}

	/* Get the protocol from the startup message.*/
	version := protocol.GetVersion(message)

	log.Printf("Client version : %v, (SSL mode: %v)", version, version == protocol.SSLRequestCode)

	/* Handle the case where the startup message was an SSL request. */
	if version == protocol.SSLRequestCode {
		return nil, fmt.Errorf("SSL not supported")
	}

	/* Now read the startup parameters */
	startup := protocol.NewMessageBuffer(message[8:length])

	var options = make(map[string]string)
	for {
		param, err := startup.ReadString()
		value, err := startup.ReadString()
		if err != nil || param == "\x00" {
			break
		}

		options[param] = value
	}

	return options, nil
}

func promptForPassword(client net.Conn) ([]byte, error) {
	message := protocol.NewMessageBuffer([]byte{})

	/* Set the message type */
	message.WriteByte(protocol.AuthenticationMessageType)

	/* Temporarily set the message length to 0. */
	message.WriteInt32(0)

	/* Set the protocol version. */
	message.WriteInt32(protocol.AuthenticationClearText)

	/* Update the message length */
	message.ResetLength(protocol.PGMessageLengthOffset)

	// Send the password message to the backend.
	_, err := connect.Send(client, message.Bytes())

	if err != nil {
		return nil, err
	}

	response := make([]byte, 4096)

	_, err = client.Read(response)
	if err != nil {
		return nil, err
	}

	message = protocol.NewMessageBuffer(response)

	code, err := message.ReadByte()
	if err != nil {
		return nil, err
	}
	if code != protocol.PasswordMessageType {
		return nil, fmt.Errorf("Expected message %d in response to password prompt, got %d", protocol.PasswordMessageType, code)
	}

	length, err := message.ReadInt32()
	if err != nil {
		return nil, err
	}

	password, err := message.ReadBytes(int(length))
	if err != nil {
		return nil, err
	}

	password = bytes.Trim(password, "\x00")
	return password, nil
}

func loadBackendConfigFromConjur(resource string) (*config.BackendConfig, error) {
	var err error
	var token *string
	var url string

	if token, err = conjur.Authenticate(HostUsername, HostAPIKey); err != nil {
		return nil, err
	}

	configuration := config.BackendConfig{}
	resourceTokens := strings.SplitN(resource, ":", 3)
	baseName := strings.Join([]string{ resourceTokens[0], "variable", resourceTokens[2] }, "/")
	if configuration.Username, err = conjur.Secret(fmt.Sprintf("%s/username", baseName), *token); err != nil {
		return nil, err
	}
	if configuration.Password, err = conjur.Secret(fmt.Sprintf("%s/password", baseName), *token); err != nil {
		return nil, err
	}
	if url, err = conjur.Secret(fmt.Sprintf("%s/url", baseName), *token); err != nil {
		return nil, err
	}

	// Form of url is : 'dbcluster.myorg.com:5432/reports'
	tokens := strings.SplitN(url, "/", 2)
	configuration.Address = tokens[0]
	if len(tokens) == 2 {
		configuration.Database = tokens[1]
	}
	configuration.Options = make(map[string]string)

	return &configuration, nil
}

func authorizeWithConjur(resource, token string) error {
	allowed, err := conjur.CheckPermission(resource, token)
	if allowed {
		return nil
	} else {
		return err
	}
}

func authenticateWithPassword(password, expectedPassword string) error {
	valid := (string(password) == expectedPassword)
	if valid {
		log.Print("Password is valid")
		return nil
	} else {
		return fmt.Errorf("Password is invalid")
	}
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
		time.Sleep(500 * time.Millisecond)
	}
}

func pipe(client net.Conn, backend net.Conn) {
	log.Printf("Connecting client %s to backend %s", client.RemoteAddr(), backend.RemoteAddr())

	go stream(client, backend)
	go stream(backend, client)
}

func connectToBackend(configuration *config.BackendConfig, database string) ([]byte, net.Conn, error) {
	connection, err := connect.Connect(configuration.Address)
	if err != nil {
		return nil, nil, err
	}

	log.Print("Sending startup message")
	startupMessage := protocol.CreateStartupMessage(configuration.Username, database, configuration.Options)

	connection.Write(startupMessage)

	response := make([]byte, 4096)
	connection.Read(response)

	log.Print("Authenticating")
	message, authenticated := connect.HandleAuthenticationRequest(configuration.Username, configuration.Password, connection, response)

	if !authenticated {
		return nil, nil, fmt.Errorf("Authentication failed")
	}

	log.Printf("Successfully connected to '%s'", configuration.Address)
	return message, connection, nil
}

func HandleConnection(configuration config.Config, client net.Conn) (error) {
	options, err := initialize(client)
	if err != nil {
		return err
	}

	log.Printf("Client options : %s", options)

	clientUser, ok := options["user"]
	if !ok {
		return fmt.Errorf("No 'user' found in connect options")
	}
	database, ok := options["database"]
	if !ok {
		return fmt.Errorf("No 'database' found in connect options")
	}

	// Authenticate and authorize with Conjur
	clientPassword, err := promptForPassword(client)
	if err != nil {
		return err
	}

	staticPassword, staticAuth := configuration.AuthorizedUsers[clientUser]
	var authenticationError error
	if staticAuth {
		// There's a statically configured password
		authenticationError = authenticateWithPassword(string(clientPassword), staticPassword)
	} else {
		log.Printf("Password for '%s' not found in static configuration. Attempting Conjur authorization.", clientUser)

		token, err := base64.StdEncoding.DecodeString(string(clientPassword))
		if err != nil {
			return err
		}
		authenticationError = authorizeWithConjur(configuration.Authorization.Resource, string(token))
	}

	if authenticationError != nil {
		if options["application_name"] == "psql" && authenticationError == io.EOF {
			log.Printf("Got %s from psql, this is normal", err)
		} else {
			log.Print(authenticationError)
			var msg string
			if staticAuth {
				msg = "Login failed"
			} else {
				msg = "Conjur authorization failed"
			}
			pgError := protocol.Error{
				Severity: protocol.ErrorSeverityFatal,
				Code:     protocol.ErrorCodeInvalidPassword,
				Message:  msg,
			}
			connect.Send(client, pgError.GetMessage())
		}
		return nil
	}

	var backendConfig *config.BackendConfig
	if staticAuth {
		backendConfig = &configuration.Backend
	} else {
		backendConfig, err = loadBackendConfigFromConjur(configuration.Authorization.Resource)
		if err != nil {
			return err
		}
	}

	msg, backend, err := connectToBackend(backendConfig, database)
	if err != nil {
		return err
	}

	_, err = connect.Send(client, msg)

	if err != nil {
		return err
	}

	pipe(client, backend)
	return nil
}

func handleConnection(configuration config.Config, client net.Conn) {
	err := HandleConnection(configuration, client)
	if err != nil {
		pgError := protocol.Error{
			Severity: protocol.ErrorSeverityFatal,
			Code:     protocol.ErrorCodeInternalError,
			Message:  err.Error(),
		}
		connect.Send(client, pgError.GetMessage())
	}
}

func Serve(configuration config.Config, l net.Listener) error {
	log.Printf("Server listening on: %s", l.Addr())

	for {
		conn, err := l.Accept()

		if err != nil {
			continue
		}

		go handleConnection(configuration, conn)
	}
}

func start(configuration config.Config) {
	proxyListener, err := net.Listen("tcp", configuration.Address)
	if err == nil {
		Serve(configuration, proxyListener)
	} else {
		log.Fatal(err)
	}
}

func main() {
	log.Println("Secretless-Postgres Starting...")

	configFile := flag.String("config", "config.yaml", "Configuration file name")
	flag.Parse()

	configuration := config.Configure(*configFile)
	log.Printf("Loaded configuration : %v", configuration)
	start(configuration)
}
