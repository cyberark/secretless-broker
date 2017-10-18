package main

import (
  "net"
  "net/http"
  "strings"
  "log"
  "fmt"
  "bytes"
  "io"
  "time"
  "flag"

  "github.com/cyberark/conjur-api-go/conjurapi"
  "github.com/kgilpin/secretless-pg/protocol"
  "github.com/kgilpin/secretless-pg/connect"
  "github.com/kgilpin/secretless-pg/config"
)

func initialize(client net.Conn) (map[string] string, error) {
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

  var options = make(map[string] string)
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

func authenticateWithConjur(resource string, client net.Conn) error {
  password, err := promptForPassword(client)
  if ( err != nil ) {
    return err
  }

  conjur_config := conjurapi.LoadConfig()

  httpClient := &http.Client{Timeout: time.Second * 10}

  resourceTokens := strings.Split(resource, ":")

  checkURL := fmt.Sprintf("%s/resources/%s?check=true&privilege=execute", conjur_config.ApplianceURL, strings.Join(resourceTokens, "/"))
  req, err := http.NewRequest(
    "GET",
    checkURL,
    nil,
  )
  if ( err != nil ) {
    return err    
  }

  req.Header.Set(
    "Authorization",
    fmt.Sprintf("Token token=\"%s\"", password),
  )

  resp, err := httpClient.Do(req)
  if err != nil {
    return err
  }

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    log.Printf("User is authorized to 'execute' %s", resource)
    return nil
  } else {
    log.Printf("Permission check failed with HTTP status %d", resp.StatusCode)
    return fmt.Errorf("Permission denied")
  }
}

func authenticateWithPassword(username string, expectedPassword string, client net.Conn) error {
  password, err := promptForPassword(client)
  if ( err != nil ) {
    return err
  }

  valid := (string(password) == expectedPassword)

  if valid {
    log.Print("Password is valid")

    return nil
  } else {
    log.Print("Password is invalid")

    pgError := protocol.Error{
      Severity: protocol.ErrorSeverityFatal,
      Code:     protocol.ErrorCodeInvalidPassword,
      Message:  "Login failed",
    }

    connect.Send(client, pgError.GetMessage())

    return fmt.Errorf("Password invalid")
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

func connectToBackend(configuration config.Config) ([]byte, net.Conn, error) {
  address := configuration.Backend.Address
  log.Printf("Connecting to backend %s...", address)
  connection, err := connect.Connect(address)

  username := configuration.Backend.Username
  password := configuration.Backend.Password
  database := configuration.Backend.Database
  options  := configuration.Backend.Options

  log.Print("Sending startup message")
  startupMessage := protocol.CreateStartupMessage(username, database, options)

  connection.Write(startupMessage)

  response := make([]byte, 4096)
  connection.Read(response)

  log.Print("Authenticating")
  message, authenticated := connect.HandleAuthenticationRequest(username, password, connection, response)

  if !authenticated {
    return nil, nil, fmt.Errorf("Authentication failed")
  }

  if err != nil {
    log.Printf("Error establishing connection to '%s'", address)
    log.Printf("Error: %s", err.Error())
    return nil, nil, fmt.Errorf("Error establishing connection to '%s'", address)
  } else {
    log.Printf("Successfully connected to '%s'", address)
    return message, connection, nil
  }
}

func HandleConnection(configuration config.Config, client net.Conn) {
  options, err := initialize(client)
  if err != nil {
    log.Print(err)
    return
  }

  log.Printf("Client options : %s", options)

  username, ok := options["user"]
  if !ok {
    log.Printf("No 'user' found in connect options")
    return
  }
  password, ok := configuration.AuthorizedUsers[username]
  var authenticationError error
  if ok {
    // There's a statically configured password
    authenticationError = authenticateWithPassword(username, password, client)
  } else {
    log.Printf("Password for '%s' not found in static configuration. Attempting Conjur authorization.", username)
    // Authenticate and authorize with Conjur
    authenticationError = authenticateWithConjur(configuration.Authorization.Resource, client)
  }

  if authenticationError != nil {
    if options["application_name"] == "psql" && authenticationError == io.EOF {
      log.Printf("Got %s from psql, this is normal", err)
    } else {
      sqlError := protocol.Error{Severity: protocol.ErrorSeverityFatal, Code: "28P01", Message: "Conjur authentication failed"}
      errorMessage := sqlError.GetMessage()
      connect.Send(client, errorMessage)
    }
    return
  }

  msg, backend, err := connectToBackend(configuration)
  if err != nil {
    log.Print(err)
    return
  }

  _, err = connect.Send(client, msg)

  if err != nil {
    log.Print(err)
    return
  }

  pipe(client, backend)
}


func Serve(configuration config.Config, l net.Listener) error {
  log.Printf("Server listening on: %s", l.Addr())

  for {
    conn, err := l.Accept()

    if err != nil {
      continue
    }

    go HandleConnection(configuration, conn)
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
