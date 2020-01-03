package mysql

// Packet represents a MySQL packet, and currently exists only to hide low level
// details about packing and unpacking the sequence ids that exist in the
// header of all MySQL protocol packets.
//
// TODO: This will be moved to protocol when we clean up that package.
type Packet []byte

// SequenceID returns the sequence id of the MySQL protocol, which is stateful.
// It allows tracking the different stages of authentication process.
func (pkt *Packet) SequenceID() byte {
	return (*pkt)[3]
}

// SetSequenceID lets Secretless set the sequence id of the MySQL protocol,
// so it can advance through the stages of the authentication process.
func (pkt *Packet) SetSequenceID(id byte) {
	(*pkt)[3] = id
}
