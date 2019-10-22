package mysql

import (
	"net"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mysql/protocol"
)

// Connection represents the entire process of sending bytes back and forth
// between a MySQL server and a MySQL client during the MySQL protocol, which
// is stateful.
//
// Since the sequence id is, conceptually, a property of *this entire process*,
// -- and not of any particular packet, even though each packet does include
// the informationa -- it makes sense that its a property of Connection.
//
// Importantly, putting the sequence id, which is part of the *header*, in
// Connection allows it to be transparent to the code that is writing the packet
// *payloads*.
type Connection struct {
	conn       net.Conn
	sequenceID byte
}

// NewClientConnection is a decorator: it takes a raw net.Conn and returns
// a mysql.Connection -- that is, a connection specific to a client talking
// to a server using the MySQL protocol.
func NewClientConnection(conn net.Conn) *Connection {
	return &Connection{sequenceID: 0, conn: conn}
}

// NewBackendConnection is a decorator: it takes a raw net.Conn and returns
// a mysql.Connection -- that is, a connection specific to a MySQL server
// (backend) talking to a client using the MySQL protocol.
func NewBackendConnection(conn net.Conn) *Connection {
	return &Connection{sequenceID: 1, conn: conn}
}

// RawConnection return the underlying net.Conn connection that
// mysql.Connection wraps.
func (c *Connection) RawConnection() net.Conn {
	return c.conn
}

// SetConnection sets the underlying net.Conn connection that
// mysql.Connection wraps.
func (c *Connection) SetConnection(conn net.Conn) {
	c.conn = conn
}

func (c *Connection) write(pkt Packet) error {
	pkt.SetSequenceID(c.sequenceID)
	c.sequenceID++
	if _, err := protocol.WritePacket(pkt, c.conn); err != nil {
		return err
	}
	return nil
}

func (c *Connection) read() (Packet, error) {
	pkt, err := protocol.ReadPacket(c.conn)
	if err != nil {
		return nil, err
	}

	mysqlPkt := Packet(pkt)
	c.sequenceID = (&mysqlPkt).SequenceID()
	c.sequenceID++
	return pkt, nil
}
