/*
MIT License

Copyright (c) 2017 Aleksandr Fedotov

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package protocol

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidPacketLength is for invalid packet lengths
var ErrInvalidPacketLength = errors.New("Protocol: Invalid packet length")

// ErrInvalidPacketType is for invalid packet types
var ErrInvalidPacketType = errors.New("Protocol: Invalid packet type")

// ErrFieldTypeNotImplementedYet is for field types that are not yet implemented
var ErrFieldTypeNotImplementedYet = errors.New("Protocol: Required field type not implemented yet")

// UnpackErrResponse decodes ERR_Packet from server.
// Part of basic packet structure shown below.
//
// int<3> PacketLength
// int<1> PacketNumber
// int<1> PacketType (0xFF)
// int<2> ErrorCode
// if clientCapabilities & clientProtocol41
//
//	{
//			string<1> SqlStateMarker (#)
//			string<5> SqlState
//	}
//
// string<EOF> Error
func UnpackErrResponse(data []byte) error {
	// Min packet length =
	// header(4 bytes)
	// + PacketType(1 byte)
	// + ErrorCode(2 bytes)
	// + string<EOF>(at least 1 byte)
	if err := CheckPacketLength(8, data); err != nil {
		return err
	}
	pos := 0

	// skip header
	pos = pos + 4

	// skip PacketType
	// 0xff [1 byte]
	pos++

	// Error Number [16 bit uint]
	errno := binary.LittleEndian.Uint16(data[pos : pos+2])
	pos = pos + 2

	sqlstate := ""
	// SQL State [optional: # + 5bytes string]
	if data[pos] == '#' {
		pos++

		sqlstate = string(data[pos : pos+5])
		pos = pos + 5
	}

	// Error Message [string]
	return Error{
		Code:     errno,
		SQLState: sqlstate,
		Message:  string(data[pos:]),
	}
}

// GetPacketType extracts the PacketType byte
// Part of basic packet structure shown below.
//
//	int<3> PacketLength
//	int<1> PacketNumber
//	int<1> PacketType (0xFF)
//	... more ...
func GetPacketType(packet []byte) byte {
	return packet[4]
}

// OkResponse represents packet sent from the server to the client to signal successful completion of a command
// https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_basic_ok_packet.html
type OkResponse struct {
	PacketType   byte
	AffectedRows uint64
	LastInsertID uint64
	StatusFlags  uint16
	Warnings     uint16
}

// UnpackOkResponse decodes OK_Packet from server.
// Part of basic packet structure shown below.
//
// int<3> PacketLength
// int<1> PacketNumber
// int<1> PacketType (0x00 or 0xFE)
// int<lenenc> AffectedRows
// int<lenenc> LastInsertID
// ... more ...
func UnpackOkResponse(packet []byte) (*OkResponse, error) {

	// Min packet length = header(4 bytes) + PacketType(1 byte)
	if err := CheckPacketLength(5, packet); err != nil {
		return nil, err
	}

	r := bytes.NewReader(packet)

	// Skip packet header
	if _, err := GetPacketHeader(r); err != nil {
		return nil, err
	}

	// Read header, validate OK
	packetType, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if packetType != ResponseOk {
		return nil, errors.New("Malformed packet")
	}

	// Read affected rows (expected value: 0 for auth)
	affectedRows, err := ReadLenEncodedInteger(r)
	if err != nil {
		return nil, err
	}

	// Read last insert ID (expected value: 0 for auth)
	lastInsertID, err := ReadLenEncodedInteger(r)
	if err != nil {
		return nil, err
	}

	// Read status flags
	statusBuf := make([]byte, 2)
	if _, err := r.Read(statusBuf); err != nil {
		return nil, err
	}
	status := binary.LittleEndian.Uint16(statusBuf)

	// Read warnings
	warningsBuf := make([]byte, 2)
	if _, err := r.Read(warningsBuf); err != nil {
		return nil, err
	}
	warnings := binary.LittleEndian.Uint16(warningsBuf)

	return &OkResponse{
		PacketType:   packetType,
		AffectedRows: affectedRows,
		LastInsertID: lastInsertID,
		StatusFlags:  status,
		Warnings:     warnings}, nil
}

// HandshakeV10 represents sever's initial handshake packet
// See https://mariadb.com/kb/en/mariadb/1-connecting-connecting/#initial-handshake-packet
type HandshakeV10 struct {
	ProtocolVersion    byte
	SequenceID         uint8
	ServerVersion      string
	ConnectionID       uint32
	StatusFlags        uint16
	CharacterSet       uint8
	ServerCapabilities uint32
	AuthPlugin         string
	Salt               []byte
}

// UnpackHandshakeV10 decodes initial handshake request from server.
// Basic packet structure shown below.
// See http://imysql.com/mysql-internal-manual/connection-phase-packets.html#packet-Protocol::HandshakeV10
//
// int<3> PacketLength
// int<1> PacketNumber
// int<1> ProtocolVersion
// string<NUL> ServerVersion
// int<4> ConnectionID
// string<8> AuthPluginDataPart1 (authentication seed)
// string<1> Reserved (always 0x00)
// int<2> ServerCapabilities (1st part)
// int<1> ServerDefaultCollation
// int<2> StatusFlags
// int<2> ServerCapabilities (2nd part)
// if capabilities & clientPluginAuth
//
//	{
//			int<1> AuthPluginDataLength
//	}
//
// else
//
//	{
//			int<1> 0x00
//	}
//
// string<10> Reserved (all 0x00)
// if capabilities & clientSecureConnection
//
//	{
//			string[$len] AuthPluginDataPart2 ($len=MAX(13, AuthPluginDataLength - 8))
//	}
//
// if capabilities & clientPluginAuth
//
//	{
//			string[NUL] AuthPluginName
//	}
func UnpackHandshakeV10(packet []byte) (*HandshakeV10, error) {
	r := bytes.NewReader(packet)

	// Header
	header, err := GetPacketHeader(r)
	if err != nil {
		return nil, err
	}

	// Read ProtocolVersion
	protoVersion, _ := r.ReadByte()

	// Read ServerVersion
	serverVersion := ReadNullTerminatedString(r)

	// Read ConnectionID
	connectionIDBuf := make([]byte, 4)
	if _, err := r.Read(connectionIDBuf); err != nil {
		return nil, err
	}
	connectionID := binary.LittleEndian.Uint32(connectionIDBuf)

	// Read AuthPluginDataPart1
	var salt []byte
	salt8 := make([]byte, 8)
	if _, err := r.Read(salt8); err != nil {
		return nil, err
	}
	salt = append(salt, salt8...)

	// Skip filler
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}

	// Read ServerCapabilities
	serverCapabilitiesLowerBuf := make([]byte, 2)
	if _, err := r.Read(serverCapabilitiesLowerBuf); err != nil {
		return nil, err
	}

	// Read ServerCharacterSet and StatusFlags
	serverCharacterSet, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	serverStatusFlagsBuf := make([]byte, 2)
	if _, err := r.Read(serverStatusFlagsBuf); err != nil {
		return nil, err
	}
	serverStatusFlags := binary.LittleEndian.Uint16(serverStatusFlagsBuf)

	// Read ExServerCapabilities
	serverCapabilitiesHigherBuf := make([]byte, 2)
	if _, err := r.Read(serverCapabilitiesHigherBuf); err != nil {
		return nil, err
	}

	// Compose ServerCapabilities from 2 bufs
	var serverCapabilitiesBuf []byte
	serverCapabilitiesBuf = append(serverCapabilitiesBuf, serverCapabilitiesLowerBuf...)
	serverCapabilitiesBuf = append(serverCapabilitiesBuf, serverCapabilitiesHigherBuf...)
	serverCapabilities := binary.LittleEndian.Uint32(serverCapabilitiesBuf)

	// Get length of AuthnPluginDataPart2
	// or read in empty byte if not included
	var authPluginDataLength byte
	if serverCapabilities&ClientPluginAuth > 0 {
		var err error
		authPluginDataLength, err = r.ReadByte()
		if err != nil {
			return nil, err
		}
	} else {
		if _, err := r.ReadByte(); err != nil {
			return nil, err
		}
	}

	// Skip reserved (all 0x00)
	if _, err := r.Seek(10, io.SeekCurrent); err != nil {
		return nil, err
	}

	// Get AuthnPluginDataPart2
	var numBytes int
	if serverCapabilities&ClientSecureConnection != 0 {
		numBytes = int(authPluginDataLength) - 8
		if numBytes < 0 || numBytes > 13 {
			numBytes = 13
		}

		salt2 := make([]byte, numBytes)
		if _, err := r.Read(salt2); err != nil {
			return nil, err
		}

		// the last byte has to be 0, and is not part of the data
		if salt2[numBytes-1] != 0 {
			return nil, errors.New("Malformed packet")
		}
		salt = append(salt, salt2[:numBytes-1]...)
	}

	var authPlugin string
	if serverCapabilities&ClientPluginAuth != 0 {
		authPlugin = ReadNullTerminatedString(r)
	}

	return &HandshakeV10{
		SequenceID:         header[3],
		ProtocolVersion:    protoVersion,
		ServerVersion:      serverVersion,
		ConnectionID:       connectionID,
		ServerCapabilities: serverCapabilities,
		AuthPlugin:         authPlugin,
		Salt:               salt,
		StatusFlags:        serverStatusFlags,
		CharacterSet:       serverCharacterSet,
	}, nil
}

// PackHandshakeV10 takes in a HandshakeResponse41 object and
// returns a handshake response packet
func PackHandshakeV10(serverHandshake *HandshakeV10) ([]byte, error) {
	// Create a buffer to write the packet data
	buffer := new(bytes.Buffer)

	// Write ProtocolVersion (int<1>)
	binary.Write(buffer, binary.LittleEndian, serverHandshake.ProtocolVersion)

	// Write ServerVersion (string<NUL>)
	buffer.WriteString(serverHandshake.ServerVersion)
	buffer.WriteByte(0)

	// Write ConnectionID (int<4>)
	binary.Write(buffer, binary.LittleEndian, serverHandshake.ConnectionID)

	// Write AuthPluginDataPart1 (string<8>)
	buffer.Write(serverHandshake.Salt[:8]) // Write the first 8 bytes of the salt

	// Write Reserved (int<1>)
	buffer.WriteByte(0)

	// Write ServerCapabilities (int<2>)
	binary.Write(buffer, binary.LittleEndian, uint16(serverHandshake.ServerCapabilities&0xFFFF)) // #nosec

	// Write ServerCharacterSet (int<1>)
	buffer.WriteByte(serverHandshake.CharacterSet)

	// Write StatusFlags (int<2>)
	binary.Write(buffer, binary.LittleEndian, serverHandshake.StatusFlags)

	// Write ServerCapabilities (int<2>), the higher part
	binary.Write(buffer, binary.LittleEndian, uint16(serverHandshake.ServerCapabilities>>16)) // #nosec

	// Write AuthPluginDataLength (int<1>) if required
	if serverHandshake.ServerCapabilities&ClientPluginAuth > 0 {
		buffer.WriteByte(byte(len(serverHandshake.Salt) + 1))
	}

	// Write Reserved (string<10>)
	buffer.Write(make([]byte, 10))

	// Calculate the length of AuthPluginDataPart2
	var authPluginDataLength byte
	if serverHandshake.ServerCapabilities&ClientSecureConnection != 0 {
		numBytes := len(serverHandshake.Salt) - 8
		if numBytes > 13 {
			numBytes = 13
		}
		authPluginDataLength = byte(numBytes)
	}

	// Write AuthPluginDataPart2 (string[$len]) if required
	if serverHandshake.ServerCapabilities&ClientSecureConnection != 0 {
		buffer.Write(serverHandshake.Salt[8 : 8+int(authPluginDataLength)])
		buffer.WriteByte(0)
	}

	// Write AuthPluginName (string<NUL>) if required
	if serverHandshake.ServerCapabilities&ClientPluginAuth > 0 {
		buffer.WriteString(serverHandshake.AuthPlugin)
		buffer.WriteByte(0)
	}

	return AddHeaderToPacket(serverHandshake.SequenceID, buffer.Bytes()), nil
}

// RemoveSSLFromHandshakeV10 removes Client SSL Capability from Server
// Handshake Packet.  Secretless needs to do this to force the client to
// communicate with Secretless without using SSL.  That half of the connection
// is insecure by design.  Secretless then (usually) adds SSL for the other
// half of the communication -- between Secretless and the MySQL server.
func RemoveSSLFromHandshakeV10(packet []byte) ([]byte, error) {
	r := bytes.NewReader(packet)
	initialLen := r.Len()

	// Skip packet header
	if _, err := GetPacketHeader(r); err != nil {
		return nil, err
	}

	// Read ProtocolVersion
	r.ReadByte()

	// Read ServerVersion
	ReadNullTerminatedString(r)

	// Read ConnectionID
	connectionIDBuf := make([]byte, 4)
	if _, err := r.Read(connectionIDBuf); err != nil {
		return nil, err
	}

	// Read AuthPluginDataPart1
	var salt []byte
	salt8 := make([]byte, 8)
	if _, err := r.Read(salt8); err != nil {
		return nil, err
	}
	salt = append(salt, salt8...)

	// Skip filler
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}

	serverCapabilitiesIndex := initialLen - r.Len()
	// Read ServerCapabilities
	serverCapabilitiesLowerBuf := make([]byte, 2)
	if _, err := r.Read(serverCapabilitiesLowerBuf); err != nil {
		return nil, err
	}

	// Skip ServerDefaultCollation and StatusFlags
	if _, err := r.Seek(3, io.SeekCurrent); err != nil {
		return nil, err
	}

	// Read ExServerCapabilities
	exServerCapabilitiesIndex := initialLen - r.Len()
	serverCapabilitiesHigherBuf := make([]byte, 2)
	if _, err := r.Read(serverCapabilitiesHigherBuf); err != nil {
		return nil, err
	}

	newPacket := make([]byte, len(packet))
	copy(newPacket, packet)

	// Compose ServerCapabilities from 2 bufs
	var serverCapabilitiesBuf []byte
	serverCapabilitiesBuf = append(serverCapabilitiesBuf, serverCapabilitiesLowerBuf...)
	serverCapabilitiesBuf = append(serverCapabilitiesBuf, serverCapabilitiesHigherBuf...)
	serverCapabilities := binary.LittleEndian.Uint32(serverCapabilitiesBuf)

	// Remove ClientSSL from serverCapabilities
	serverCapabilities = serverCapabilities ^ ClientSSL

	// update Lower part of the capability flags.
	writeUint16(newPacket, serverCapabilitiesIndex, uint16(serverCapabilities)) // #nosec

	// update Upper part of the capability flags.
	writeUint16(newPacket, exServerCapabilitiesIndex, uint16(serverCapabilities>>16)) // #nosec

	return newPacket, nil
}

// writes Uint16 starting from a given position in a byte slice
func writeUint16(data []byte, pos int, value uint16) {
	data[pos] = byte(value)
	data[pos+1] = byte(value >> 8)
}

// HandshakeResponse41 represents handshake response packet sent by 4.1+ clients supporting clientProtocol41 capability,
// if the server announced it in its initial handshake packet.
// See https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase_packets_protocol_handshake_response.html#sect_protocol_connection_phase_packets_protocol_handshake_response41
//
// The format of the header is also described here:
//
//	  https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_basic_packets.html
//
//	 +-------------+----------------+---------------------------------------------+
//	 |    Type     |      Name      |                 Description                 |
//		+-------------+----------------+---------------------------------------------+
//	 | int<3>      | payload_length | Length of the payload. The number of bytes  |
//	 |             |                | in the packet beyond the initial 4 bytes    |
//	 |             |                | that make up the packet header.             |
//	 | int<1>      | sequence_id    | Sequence ID                                 |
//	 | string<var> | payload        | [len=payload_length] payload of the packet  |
//	 +-------------+----------------+---------------------------------------------+
type HandshakeResponse41 struct {
	SequenceID      uint8
	CapabilityFlags uint32
	MaxPacketSize   uint32
	ClientCharset   uint8
	Username        string
	AuthLength      int64
	AuthPluginName  string
	AuthResponse    []byte
	Database        string
	PacketTail      []byte
}

// UnpackHandshakeResponse41 decodes handshake response packet send by client.
// TODO: Add packet struct comment
// TODO: Add packet length check
func UnpackHandshakeResponse41(packet []byte) (*HandshakeResponse41, error) {
	r := bytes.NewReader(packet)

	// Skip packet header (but save in struct)
	header, err := GetPacketHeader(r)
	if err != nil {
		return nil, err
	}

	// Read CapabilityFlags
	clientCapabilitiesBuf := make([]byte, 4)
	if _, err := r.Read(clientCapabilitiesBuf); err != nil {
		return nil, err
	}
	capabilityFlags := binary.LittleEndian.Uint32(clientCapabilitiesBuf)

	// Check that the server is using protocol 4.1
	if capabilityFlags&ClientProtocol41 == 0 {
		return nil, errors.New("Client Protocol mismatch")
	}

	// client requesting SSL, we don't support it
	clientRequestedSSL := capabilityFlags&ClientSSL > 0
	if clientRequestedSSL {
		return nil, errors.New("SSL Protocol mismatch")
	}

	// Read MaxPacketSize
	maxPacketSizeBuf := make([]byte, 4)
	if _, err := r.Read(maxPacketSizeBuf); err != nil {
		return nil, err
	}
	maxPacketSize := binary.LittleEndian.Uint32(maxPacketSizeBuf)

	// Read Charset
	charset, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	// Skip 23 byte buffer
	if _, err := r.Seek(23, io.SeekCurrent); err != nil {
		return nil, err
	}

	// Read Username
	username := ReadNullTerminatedString(r)

	// Read Auth
	var auth []byte
	var authLength int64
	if capabilityFlags&ClientSecureConnection > 0 {
		authLengthByte, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		authLength = int64(authLengthByte)

		auth = make([]byte, authLength)
		if _, err := r.Read(auth); err != nil {
			return nil, err
		}
	} else {
		auth = ReadNullTerminatedBytes(r)
	}

	// Read Database
	var database string
	if capabilityFlags&ClientConnectWithDB > 0 {
		database = ReadNullTerminatedString(r)
	}

	// check whether the auth method was specified
	var authPluginName string
	if capabilityFlags&ClientPluginAuth > 0 {
		authPluginName = ReadNullTerminatedString(r)
	}

	// get the rest of the packet
	var packetTail []byte
	remainingByteLen := r.Len()
	if remainingByteLen > 0 {
		packetTail = make([]byte, remainingByteLen)
		if _, err := r.Read(packetTail); err != nil {
			return nil, err
		}
	}

	return &HandshakeResponse41{
		SequenceID:      header[3],
		CapabilityFlags: capabilityFlags,
		MaxPacketSize:   maxPacketSize,
		ClientCharset:   charset,
		Username:        username,
		AuthLength:      authLength,
		AuthPluginName:  authPluginName,
		AuthResponse:    auth,
		Database:        database,
		PacketTail:      packetTail}, nil
}

// CreateAuthResponse creates an auth response for the given auth plugin
func CreateAuthResponse(authPlugin string, password []byte, salt []byte) ([]byte, error) {
	var authResponse []byte
	var err error

	switch authPlugin {
	case "mysql_native_password":
		authResponse, err = NativePassword([]byte(password), salt)
	case "caching_sha2_password":
		authResponse = scrambleSHA256Password([]byte(password), salt)
	default:
		err = fmt.Errorf("Unknown auth plugin: %s", authPlugin)
	}

	if err != nil {
		return nil, err
	}
	return authResponse, nil
}

// InjectCredentials takes in a HandshakeResponse41 from the client, the
// salt from the server, and a username / password, and uses the salt
// from the server handshake to inject the username / password credentials into
// the client handshake response
func InjectCredentials(authPlugin string, clientHandshake *HandshakeResponse41, salt []byte, username string, password string) (err error) {
	authResponse, err := CreateAuthResponse(authPlugin, []byte(password), salt)
	if err != nil {
		return
	}

	clientHandshake.AuthPluginName = authPlugin
	clientHandshake.Username = username
	clientHandshake.AuthLength = int64(len(authResponse))
	clientHandshake.AuthResponse = authResponse

	return
}

// Hash password using MySQL 8+ method (SHA256)
func scrambleSHA256Password(password []byte, scramble []byte) []byte {
	if len(password) == 0 {
		return nil
	}

	// XOR(SHA256(password), SHA256(SHA256(SHA256(password)), scramble))

	crypt := sha256.New()
	crypt.Write(password)
	message1 := crypt.Sum(nil)

	crypt.Reset()
	crypt.Write(message1)
	message1Hash := crypt.Sum(nil)

	crypt.Reset()
	crypt.Write(message1Hash)
	crypt.Write(scramble)
	message2 := crypt.Sum(nil)

	for i := range message1 {
		message1[i] ^= message2[i]
	}

	return message1
}

// PackHandshakeResponse41 takes in a HandshakeResponse41 object and
// returns a handshake response packet
func PackHandshakeResponse41(clientHandshake *HandshakeResponse41) ([]byte, error) {

	var buf bytes.Buffer

	// write the capability flags
	capabilityFlagsBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(capabilityFlagsBuf, clientHandshake.CapabilityFlags)
	buf.Write(capabilityFlagsBuf)

	// write max packet size
	maxPacketSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(maxPacketSizeBuf, clientHandshake.MaxPacketSize)
	buf.Write(maxPacketSizeBuf)

	// write 1 byte char set
	buf.WriteByte(clientHandshake.ClientCharset)

	// write string[23] reserved (all zero)
	for i := 0; i < 23; i++ {
		buf.WriteByte(0)
	}

	// write string username
	buf.WriteString(clientHandshake.Username)
	buf.WriteByte(0)

	// write auth
	if clientHandshake.CapabilityFlags&ClientSecureConnection > 0 {
		if clientHandshake.AuthLength > 0 {
			buf.WriteByte(uint8(len(clientHandshake.AuthResponse))) // #nosec
			buf.Write(clientHandshake.AuthResponse)
		} else {
			buf.WriteByte(0)
		}
	} else {
		buf.Write(clientHandshake.AuthResponse)
		buf.WriteByte(0)
	}

	// write database (if set)
	if clientHandshake.CapabilityFlags&ClientConnectWithDB > 0 {
		buf.WriteString(clientHandshake.Database)
		buf.WriteByte(0)
	}

	// write auth plugin name
	buf.WriteString(clientHandshake.AuthPluginName)
	buf.WriteByte(0)

	// write tail of packet (if set)
	if len(clientHandshake.PacketTail) > 0 {
		buf.Write(clientHandshake.PacketTail)
	}

	return AddHeaderToPacket(clientHandshake.SequenceID, buf.Bytes()), nil
}

// GetLenEncodedIntegerSize returns bytes count for length encoded integer
// determined by it's 1st byte
func GetLenEncodedIntegerSize(firstByte byte) byte {
	switch firstByte {
	case 0xfc:
		return 2
	case 0xfd:
		return 3
	case 0xfe:
		return 8
	default:
		return 1
	}
}

// ReadLenEncodedInteger returns parsed length-encoded integer and it's offset.
// See https://mariadb.com/kb/en/mariadb/protocol-data-types/#length-encoded-integers
func ReadLenEncodedInteger(r *bytes.Reader) (value uint64, err error) {
	firstLenEncIntByte, err := r.ReadByte()
	if err != nil {
		return
	}

	switch firstLenEncIntByte {
	case 0xfb:
		value = 0

	case 0xfc:
		data := make([]byte, 2)
		_, err = r.Read(data)
		if err != nil {
			return
		}
		value = uint64(data[0]) | uint64(data[1])<<8

	case 0xfd:
		data := make([]byte, 3)
		_, err = r.Read(data)
		if err != nil {
			return
		}
		value = uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16

	case 0xfe:
		data := make([]byte, 8)
		_, err = r.Read(data)
		if err != nil {
			return
		}
		value = uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16 |
			uint64(data[3])<<24 | uint64(data[4])<<32 | uint64(data[5])<<40 |
			uint64(data[6])<<48 | uint64(data[7])<<56

	default:
		value = uint64(firstLenEncIntByte)
	}

	return value, err
}

// ReadLenEncodedString returns parsed length-encoded string and it's length.
// Length-encoded strings are prefixed by a length-encoded integer which describes
// the length of the string, followed by the string value.
// See https://mariadb.com/kb/en/mariadb/protocol-data-types/#length-encoded-strings
func ReadLenEncodedString(r *bytes.Reader) (string, uint64, error) {
	strLen, _ := ReadLenEncodedInteger(r)

	strBuf := make([]byte, strLen)
	if _, err := r.Read(strBuf); err != nil {
		return "", 0, err
	}

	return string(strBuf), strLen, nil
}

// ReadEOFLengthString returns parsed EOF-length string.
// EOF-length strings are those strings whose length will be calculated by the packet remaining length.
// See https://mariadb.com/kb/en/mariadb/protocol-data-types/#end-of-file-length-strings
func ReadEOFLengthString(data []byte) string {
	return string(data)
}

// ReadNullTerminatedString reads bytes from reader until 0x00 byte
// See https://mariadb.com/kb/en/mariadb/protocol-data-types/#null-terminated-strings
func ReadNullTerminatedString(r *bytes.Reader) string {
	var str []byte
	for {
		//TODO: Check for error
		b, _ := r.ReadByte()

		if b == 0x00 {
			return string(str)
		}

		str = append(str, b)
	}
}

// AuthSwitchRequest represents a request from the server to switch to a different authentication method.
// https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase_packets_protocol_auth_switch_request.html
type AuthSwitchRequest struct {
	SequenceNumber uint8
	PluginName     string
	PluginData     []byte
}

// UnpackAuthSwitchRequest decodes an AuthSwitchRequest packet from the provided data.
func UnpackAuthSwitchRequest(data []byte) (*AuthSwitchRequest, error) {
	sequenceNumber := data[3]
	data = data[4:]

	// Find the position of the null-terminated plugin name
	nullTerminatorIndex := 1
	for i := 1; i < len(data); i++ {
		if data[i] == 0x00 {
			nullTerminatorIndex = i
			break
		}
	}

	// Extract the plugin name
	if nullTerminatorIndex == 1 {
		return nil, fmt.Errorf("Invalid AuthSwitchRequest packet: Missing plugin name")
	}
	pluginName := string(data[1:nullTerminatorIndex])

	// Extract the plugin provided data
	var pluginData []byte
	if nullTerminatorIndex+1 < len(data) {
		pluginData = data[nullTerminatorIndex+1:]
	}

	return &AuthSwitchRequest{
		SequenceNumber: sequenceNumber,
		PluginName:     pluginName,
		PluginData:     pluginData,
	}, nil
}

// UnpackAuthRequestPubKeyResponse decodes a response from the server to a request for its public key.
func UnpackAuthRequestPubKeyResponse(data []byte) (*rsa.PublicKey, error) {
	// Parse public key
	if data[4] != ResponseAuthMoreData {
		return nil, fmt.Errorf("expected ResponseAuthMoreData packet")
	}

	block, rest := pem.Decode(data[5:])
	if block == nil {
		return nil, fmt.Errorf("no pem data found, data: %s", rest)
	}
	pkix, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %s", err)
	}
	return pkix.(*rsa.PublicKey), nil
}

// ReadNullTerminatedBytes reads bytes from reader until 0x00 byte
func ReadNullTerminatedBytes(r *bytes.Reader) (str []byte) {
	for {
		//TODO: Check for error
		b, _ := r.ReadByte()

		if b == 0x00 {
			return
		}

		str = append(str, b)
	}
}

// GetPacketHeader rewinds reader to packet payload
func GetPacketHeader(r *bytes.Reader) (s []byte, e error) {
	s = make([]byte, 4)

	if _, e = r.Read(s); e != nil {
		return nil, e
	}

	return
}

// CheckPacketLength checks if packet length meets expected value
func CheckPacketLength(expected int, packet []byte) error {
	if len(packet) < expected {
		return ErrInvalidPacketLength
	}

	return nil
}

// NativePassword calculates native password expected by server in HandshakeResponse41
// https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase_packets_protocol_handshake_response.html#sect_protocol_connection_phase_packets_protocol_handshake_response41
// SHA1( password ) XOR SHA1( "20-bytes random data from server" <concat> SHA1( SHA1( password ) ) )
func NativePassword(password []byte, salt []byte) (nativePassword []byte, err error) {
	sha1 := sha1.New()
	sha1.Write(password)
	passwordSHA1 := sha1.Sum(nil)

	sha1.Reset()
	sha1.Write(passwordSHA1)
	hash := sha1.Sum(nil)

	sha1.Reset()
	sha1.Write(salt)
	sha1.Write(hash)
	randomSHA1 := sha1.Sum(nil)

	// nativePassword = passwordSHA1 ^ randomSHA1
	nativePassword = make([]byte, len(randomSHA1))
	for i := range randomSHA1 {
		nativePassword[i] = passwordSHA1[i] ^ randomSHA1[i]
	}

	return
}

// AddHeaderToPacket adds a header to a packet
func AddHeaderToPacket(sequenceID uint8, restOfPacket []byte) []byte {
	// Calculate the packet length (excluding the length field itself)
	packetLength := len(restOfPacket)

	// Create a header buffer and write the packet length (int<3>)
	headerBuffer := make([]byte, 4)
	headerBuffer[0] = byte(packetLength & 0xFF)
	headerBuffer[1] = byte((packetLength >> 8) & 0xFF)
	headerBuffer[2] = byte((packetLength >> 16) & 0xFF)
	headerBuffer[3] = sequenceID

	// Combine the header and packet data to create the final packet
	return append(headerBuffer, restOfPacket...)
}

// PackAuthSwitchResponse creates an AuthSwitchResponse packet with the provided response data.
func PackAuthSwitchResponse(authSwitchRequestSequenceID uint8, data []byte) ([]byte, error) {
	// Create a buffer to write the packet data
	buffer := new(bytes.Buffer)

	// Write the response data to the buffer
	buffer.Write(data)

	return AddHeaderToPacket(authSwitchRequestSequenceID, buffer.Bytes()), nil
}

// AuthMoreDataResponse represents a packet sent from the server to request more auth data from the client.
type AuthMoreDataResponse struct {
	SequenceID uint8
	PacketType byte
	StatusTag  byte
}

// UnpackAuthMoreDataResponse decodes AuthMoreData from server.
// Basic packet structure shown below.
//
// int<3> PacketLength
// int<1> PacketNumber
// int<1> PacketType (0x01)
// int<1> StatusTag (0x03 or 0x04)
// string<EOF> AuthenticationMethodData (unused by secretless)
func UnpackAuthMoreDataResponse(packet []byte) (*AuthMoreDataResponse, error) {

	// Min packet length = header(4 bytes) + PacketType(1 byte)
	if err := CheckPacketLength(5, packet); err != nil {
		return nil, err
	}

	r := bytes.NewReader(packet)

	header, err := GetPacketHeader(r)
	if err != nil {
		return nil, err
	}

	// Read header, validate OK
	packetType, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if packetType != ResponseAuthMoreData {
		return nil, errors.New("Malformed packet")
	}

	// Read status tag
	statusTag, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	return &AuthMoreDataResponse{
		SequenceID: header[3],
		PacketType: packetType,
		StatusTag:  statusTag,
	}, nil
}

// PackAuthRequestPubKeyResponse encodes the request for the server's public key
func PackAuthRequestPubKeyResponse(sequenceID uint8) []byte {
	return AddHeaderToPacket(sequenceID, []byte{CachingSha2PasswordRequestPublicKey})
}

// PackAuthEncryptedPasswordResponse encodes the encrypted password response packet
func PackAuthEncryptedPasswordResponse(sequenceID uint8, encPwd []byte) []byte {
	return AddHeaderToPacket(sequenceID, encPwd)
}

// EncryptPassword encrypts a password using the provided seed and public key.
func EncryptPassword(password string, seed []byte, pub *rsa.PublicKey) ([]byte, error) {
	// For this stage of the authentication, we must use sha1. See
	// https://github.com/go-sql-driver/mysql/blob/19171b59bf90e6bf7a5bdf979e5e24a84b328b8a/auth.go#L217-L226
	plain := make([]byte, len(password)+1)
	copy(plain, password)
	for i := range plain {
		j := i % len(seed)
		plain[i] ^= seed[j]
	}
	sha1 := sha1.New()
	return rsa.EncryptOAEP(sha1, rand.Reader, pub, plain, nil)
}
