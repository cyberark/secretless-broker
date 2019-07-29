package mysql

import (
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strconv"

	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/mysql/protocol"
	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
)

// Handler connects a client to a backend. It uses the handler Config and Providers to
// establish the ConnectionDetails, which is used to make the Backend connection. Then the data
// is transferred bidirectionally between the Client and Backend.
//
// Handler requires "host", "port", "username" and "password" credentials.
//
type Handler struct {

	// The connections are decorated versions of net.Conn that allow us
	// to do read/writes according to the MySQL protocol.  Protocol level
	// details are thus encapsulated.  Within the MySQL code, we _only_
	// deal with these decorated versions.

	mySQLClientConn *Connection
	mySQLBackendConn *Connection
	plugin_v1.BaseHandler
	connectionDetails *ConnectionDetails
}

// fetchConnectionDetails looks up the provider credentials and returns them
// in a ConnectionDetails.  This almost could have been called
// "NewConnectionDetails", but it's not quite a constructor.  I used "fetch" to
// emphasize that this is not a pure contructor, but a side effectful one that
// makes a network request and can return an error.  And, of course, it's a
// method on handler.
//
func (h *Handler) fetchConnectionDetails() (result *ConnectionDetails, err error) {
	result = &ConnectionDetails{Options: make(map[string]string)}

	var connectionDetails map[string][]byte
	if connectionDetails, err = h.Resolver.Resolve(h.GetConfig().Credentials); err != nil {
		return nil, err
	}

	if h.DebugModeOn() {
		keys := reflect.ValueOf(connectionDetails).MapKeys()
		log.Printf("%s backend connection parameters: %s", h.GetConfig().Name, keys)
	}

	if host := connectionDetails["host"]; host != nil {
		result.Host = string(connectionDetails["host"])
	}

	if connectionDetails["port"] != nil {
		port64, _ := strconv.ParseUint(string(connectionDetails["port"]), 10, 64)
		result.Port = uint(port64)
	}

	if connectionDetails["username"] != nil {
		result.Username = string(connectionDetails["username"])
	}

	if connectionDetails["password"] != nil {
		result.Password = string(connectionDetails["password"])
	}

	delete(connectionDetails, "host")
	delete(connectionDetails, "port")
	delete(connectionDetails, "username")
	delete(connectionDetails, "password")

	result.Options = make(map[string]string)
	for k, v := range connectionDetails {
		result.Options[k] = string(v)
	}

	return result, nil
}

// If the error is not already a MySQL protocol error, then wrap it in an
// "Unknown" MySQL protocol error, because the client understands only those.
//
func (h *Handler) sendErrorToClient(err error) {
	mysqlErrorContainer, isProtocolErr := err.(protocol.ErrorContainer)
	if !isProtocolErr {
		mysqlErrorContainer = protocol.NewGenericError(err)
	}

	if e := h.mySQLClientConn.write(mysqlErrorContainer.GetPacket()); e != nil {
		log.Printf("Attempted to write error %s to MySQL client but failed", e)
	}
}

// TODO: This feels like shared code applicable to all handlers
//
func stream(source, dest net.Conn, callback func([]byte)) {
	defer func() {
		source.Close()
		dest.Close()
	}()

	buffer := make([]byte, 4096)

	var length int
	var readErr error
	var writeErr error

	for {
		length, readErr = source.Read(buffer)

		// Ensure the source packet is sent to the destination prior to inspecting errors
		// Q: Does this make sense?
		_, writeErr = dest.Write(buffer[:length])

		if readErr != nil {
			if readErr == io.EOF {
				log.Printf("source %s closed for destination %s", source.RemoteAddr(), dest.RemoteAddr())
			}
			return
		}

		if writeErr != nil {
			return
		}

		callback(buffer[:length])
	}
}

// pipe performs continuous bidirectional transfer of data between the client and backend.
//
func (h *Handler) pipe() {
	plugin_v1.PipeHandlerWithStream(h, stream, h.EventNotifier, func() {
		h.ShutdownNotifier(h)
	})
}

// Run is the main handler method. It:
//
//   1. Fetches credentials from the provider.
//   2. Dials the backend.
//   3. Runs through the connection phase steps to authenticate.
//   4. Pipes all future bytes unaltered between client and server.
//
func (h *Handler) Run() {

	// Upgrade to a decorated connection that handles protocol details for us
	// We need to do this first because sendErrorToClient uses this to send error messages.
	//
	h.mySQLClientConn = NewClientConnection(h.ClientConnection)

	// 1. Fetches credentials from the provider.
	//
	connDetails, err := h.fetchConnectionDetails()
	if err != nil {
		h.sendErrorToClient(err)
		return
	}

	// 2. Dials the backend.
	//
	rawBackendConn, err := net.Dial("tcp", connDetails.Address())
	if err != nil {
		h.sendErrorToClient(err)
		return
	}

	h.mySQLBackendConn = NewBackendConnection(rawBackendConn)

	// 3. Runs through the connection phase steps to authenticate.
	//
	connPhase := NewAuthenticationHandshake(h.mySQLClientConn, h.mySQLBackendConn, connDetails)

	if err = connPhase.Run(); err != nil {
		h.sendErrorToClient(err)
		return
	}

	h.BackendConnection = connPhase.AuthenticatedBackendConn() // conn may have changed

	h.Debugf("Successfully connected to '%s:%d'", connDetails.Host, connDetails.Port)

	// 4. Pipes all future bytes unaltered between client and server.
	//
	h.pipe()
}

// Shutdown tries to nicely close our connection
func (h *Handler) Shutdown() {
	defer h.BaseHandler.Shutdown()

	h.sendErrorToClient(fmt.Errorf("handler shut down by secretless"))
}

// HandlerFactory instantiates a handler given HandlerOptions
func HandlerFactory(options plugin_v1.HandlerOptions) plugin_v1.Handler {
	handler := &Handler{
		BaseHandler: plugin_v1.NewBaseHandler(options),
	}

	// TODO: This was a mistake.  The constructor should NOT be calling Run() on
	//   itself.
	handler.Run()

	return handler
}
