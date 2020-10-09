package mongodb

import (
	"context"
	"net"

	"go.mongodb.org/mongo-driver/mongo/address"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"

	"github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
)

// SingleUseConnector is passed the client's net.Conn and the current CredentialValuesById,
// and returns an authenticated net.Conn to the target service
type SingleUseConnector struct {
	logger log.Logger
}

// Connect receives a connection to the client, and opens a connection to the target using the client's connection
// and the credentials provided in credentialValuesByID
func (connector *SingleUseConnector) Connect(
	clientConn net.Conn,
	credentialValuesByID connector.CredentialValuesByID,
) (net.Conn, error) {
	//connDetails, _ := NewConnectionDetails(credentialValuesByID)

	//host := net.JoinHostPort(connDetails.Host, fmt.Sprintf("%d", connDetails.Port))
	//backendConn, err := dialer.DialContext(context.Background(), "tcp", host)
	//if err != nil {
	//	return nil, err
	////}
	//authenticator, err := auth.CreateAuthenticator("SCRAM-SHA-1", &auth.Cred{
	//	Source:      "admin",
	//	Username:    "user0",
	//	Password:    "pass0",
	//})
	//if err != nil {
	//	return nil, err
	//}

	var firstResponse []byte

	connString, err := connstring.ParseAndValidate(
		string(credentialValuesByID["connString"]),
	)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(connString.Hosts[0])
	if err != nil {
		return nil, err
	}

	conn, err := topology.NewConnection(address.Address(net.JoinHostPort(host, port)),
		topology.WithDialer(func(topology.Dialer) topology.Dialer {
			return newProxyDialer(func(bytes []byte) {
				firstResponse = bytes
			})
		}),
		topology.WithConnStringForConn(func(connstring.ConnString) connstring.ConnString {
			return connString
		}),
		//topology.WithHandshaker(func(h topology.Handshaker) topology.Handshaker {
		//	options := &auth.HandshakeOptions{
		//		AppName:               "meow",
		//		Authenticator:         authenticator,
		//		PerformAuthentication: func(server description.Server) bool {
		//			return true
		//		},
		//	}
		//
		//	return auth.Handshaker(h, options)
		//}),
	)
	if err != nil {
		return nil, err
	}

	conn.Connect(context.Background())
	err = conn.Wait()
	if err != nil {
		return nil, err
	}

	stolenConn := conn.StealConn()

	if _stolenConn, ok := stolenConn.(*proxyConn); ok {
		stolenConn = _stolenConn.NetConn()
	}

	bytes := make([]byte, 2048)
	readBytes, err := clientConn.Read(bytes)
	if err != nil {
		return nil, err
	}
	//fmt.Println("Read from the client", readBytes)
	clientFirstMessage, err := parseSentMessage(bytes[:readBytes])
	if err != nil {
		return nil, err
	}

	// Do something with the client first message
	if clientFirstMessage == clientFirstMessage {}

	err = backendTemp(
		clientConn,
		connString.Hosts[0],
		bytes,
		readBytes,
	)
	//if err != nil {
	//	fmt.Println("Warning:", "Tried to use backend temp connection and got:", err)
	//	fmt.Println("Will attempt to use first response");
	//	_, err = clientConn.Write(firstResponse)
	//}
	if err != nil {
		return nil, err
	}

	return stolenConn, nil
}

func backendTemp(
	clientConn net.Conn,
	host string,
	bytes []byte,
	readBytes int,
) error  {
	backendTmpConn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}

	_, err = backendTmpConn.Write(bytes[:readBytes])
	if err != nil {
		return err
	}

	readBytes, err = backendTmpConn.Read(bytes)
	if err != nil {
		return err
	}

	// writtenBytes, err := clientConn.Write(bytes[:readBytes])
	_, err = clientConn.Write(bytes[:readBytes])
	if err != nil {
		return err
	}

	// fmt.Println("Wrote to the client", readBytes, writtenBytes)

	return nil
}