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
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strconv"
)

var errInvalidPacketLength = errors.New("Protocol: Invalid packet length")
var errInvalidPacketType = errors.New("Protocol: Invalid packet type")
var errFieldTypeNotImplementedYet = errors.New("Protocol: Required field type not implemented yet")

type ErrResponse struct {
	Message string
}

func DecodeErrResponse(packet []byte) (*ErrResponse, error) {
	return nil, nil
}

// OkResponse represents packet sent from the server to the client to signal successful completion of a command
// https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_basic_ok_packet.html
type OkResponse struct {
	PacketType   byte
	AffectedRows uint64
	LastInsertID uint64
}

// DecodeOkResponse decodes OK_Packet from server.
// Part of basic packet structure shown below.
//
// int<3> PacketLength
// int<1> PacketNumber
// int<1> PacketType (0x00 or 0xFE)
// int<lenenc> AffectedRows
// int<lenenc> LastInsertID
// ... more ...
func DecodeOkResponse(packet []byte) (*OkResponse, error) {

	// Min packet length = header(4 bytes) + PacketType(1 byte)
	if err := CheckPacketLength(5, packet); err != nil {
		return nil, err
	}

	r := bytes.NewReader(packet)

	// Skip packet header
	if err := SkipPacketHeader(r); err != nil {
		return nil, err
	}

	// Skip packet type
	if _, err := r.Seek(1, io.SeekCurrent); err != nil {
		return nil, err
	}

	affectedRows, _ := ReadLenEncodedInteger(r)
	lastInsertID, _ := ReadLenEncodedInteger(r)

	return &OkResponse{packet[4], affectedRows, lastInsertID}, nil
}

// HandshakeV10 represents sever's initial handshake packet
// See https://mariadb.com/kb/en/mariadb/1-connecting-connecting/#initial-handshake-packet
type HandshakeV10 struct {
	ProtocolVersion    byte
	ServerVersion      string
	ConnectionID       uint32
	ServerCapabilities uint32
	AuthPlugin         string
}

// DecodeHandshakeV10 decodes initial handshake request from server.
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
// {
// 		int<1> AuthPluginDataLength
// }
// else
// {
//		int<1> 0x00
// }
// string<10> Reserved (all 0x00)
// if capabilities & clientSecureConnection
// {
// 		string[$len] AuthPluginDataPart2 ($len=MAX(13, AuthPluginDataLength - 8))
// }
// if capabilities & clientPluginAuth
// {
//		string[NUL] AuthPluginName
// }
func DecodeHandshakeV10(packet []byte) (*HandshakeV10, error) {
	// TODO: Add length check

	r := bytes.NewReader(packet)

	// Skip packet header
	if err := SkipPacketHeader(r); err != nil {
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

	// Skip AuthPluginData and filler (always 0x00)
	if _, err := r.Seek(9, io.SeekCurrent); err != nil {
		return nil, err
	}

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
	serverCapabilitiesHigherBuf := make([]byte, 2)
	if _, err := r.Read(serverCapabilitiesHigherBuf); err != nil {
		return nil, err
	}

	// Compose ServerCapabilities from 2 bufs
	var serverCapabilitiesBuf []byte
	serverCapabilitiesBuf = append(serverCapabilitiesBuf, serverCapabilitiesLowerBuf...)
	serverCapabilitiesBuf = append(serverCapabilitiesBuf, serverCapabilitiesHigherBuf...)
	serverCapabilities := binary.LittleEndian.Uint32(serverCapabilitiesBuf)

	var authPluginDataLength byte
	if serverCapabilities&clientPluginAuth != 0 {
		var err error
		authPluginDataLength, err = r.ReadByte()
		if err != nil {
			return nil, err
		}
	}

	// Skip reserved (all 0x00)
	if _, err := r.Seek(10, io.SeekCurrent); err != nil {
		return nil, err
	}

	if serverCapabilities&clientSecureConnection != 0 {
		skip := int64(math.Max(13, float64(authPluginDataLength)-8))
		// Skip reserved (all 0x00)
		if _, err := r.Seek(skip, io.SeekCurrent); err != nil {
			return nil, err
		}
	}

	var authPlugin string
	if serverCapabilities&clientPluginAuth != 0 {
		authPlugin = ReadNullTerminatedString(r)
	}

	return &HandshakeV10{
		ProtocolVersion:    protoVersion,
		ServerVersion:      serverVersion,
		ConnectionID:       connectionID,
		ServerCapabilities: serverCapabilities,
		AuthPlugin:         authPlugin,
	}, nil
}

// HandshakeResponse41 represents handshake response packet sent by 4.1+ clients supporting clientProtocol41 capability,
// if the server announced it in its initial handshake packet.
// See http://imysql.com/mysql-internal-manual/connection-phase-packets.html#packet-Protocol::HandshakeResponse41
type HandshakeResponse41 struct {
	ClientCapabilities uint32
	ClientCharset      byte
	Username           string
	Database           string
}

// DecodeHandshakeResponse41 decodes handshake response packet send by client.
// TODO: Add packet struct comment
// TODO: Add packet length check
func DecodeHandshakeResponse41(packet []byte) (*HandshakeResponse41, error) {
	r := bytes.NewReader(packet)

	// Skip packet header
	if err := SkipPacketHeader(r); err != nil {
		return nil, err
	}

	// Read CapabilityFlags
	clientCapabilitiesBuf := make([]byte, 4)
	if _, err := r.Read(clientCapabilitiesBuf); err != nil {
		return nil, err
	}
	clientCapabilities := binary.LittleEndian.Uint32(clientCapabilitiesBuf)

	// Skip MaxPacketSize
	if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return nil, err
	}

	// Read Charset
	charset, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	// Read Username
	username := ""
	for {
		userchar, err := r.ReadByte()
		if err != nil {
			return nil, err
		} else if userchar == 0 {
			break
		} else {
			username += string(userchar)
		}
	}

	// Read Auth
	var authLength int64
	if authLength, err := r.ReadByte(); err != nil {
		return nil, err
	}
	if auth, err := r.Seek(authLength, io.SeekCurrent); err != nil {
		return nil, err
	}

	// Read Database

	return &HandshakeResponse41{clientCapabilities, charset}, nil
}

// QueryRequest represents COM_QUERY or COM_STMT_PREPARE command sent by client to server.
type QueryRequest struct {
	Query string // SQL query value
}

// DecodeQueryRequest decodes COM_QUERY and COM_STMT_PREPARE requests from client.
// Basic packet structure shown below.
// See https://mariadb.com/kb/en/mariadb/com_query/ and https://mariadb.com/kb/en/mariadb/com_stmt_prepare/
//
// int<3> PacketLength
// int<1> PacketNumber
// int<1> Command COM_QUERY (0x03) or COM_STMT_PREPARE (0x16)
// string<EOF> SQLStatement
func DecodeQueryRequest(packet []byte) (*QueryRequest, error) {

	// Min packet length = header(4 bytes) + command(1 byte) + SQLStatement(at least 1 byte)
	if len(packet) < 6 {
		return nil, errInvalidPacketLength
	}

	// Fifth byte is command
	if packet[4] != ComQuery && packet[4] != ComStmtPrepare {
		return nil, errInvalidPacketType
	}

	return &QueryRequest{ReadEOFLengthString(packet[5:])}, nil
}

// ComStmtPrepareOkResponse represents COM_STMT_PREPARE_OK response structure.
type ComStmtPrepareOkResponse struct {
	StatementID   uint32 // ID of prepared statement
	ParametersNum uint16 // Num of prepared parameters
}

// DecodeComStmtPrepareOkResponse decodes COM_STMT_PREPARE_OK response from MySQL server.
// Basic packet structure shown below.
// See https://mariadb.com/kb/en/mariadb/com_stmt_prepare/#COM_STMT_PREPARE_OK
//
// int<3> PacketLength
// int<1> PacketNumber
// int<1> Command COM_STMT_PREPARE_OK (0x00)
// int<4> StatementID
// int<2> NumberOfColumns
// int<2> NumberOfParameters
// string<1> <not used>
// int<2> NumberOfWarnings
func DecodeComStmtPrepareOkResponse(packet []byte) (*ComStmtPrepareOkResponse, error) {

	// Min packet length = header(4 bytes) + command(1 byte) + statementID(4 bytes)
	// + number of columns (2 bytes) + number of parameters (2 bytes)
	// + <not used> (1 byte) + number of warnings (2 bytes)
	if len(packet) < 16 {
		return nil, errInvalidPacketLength
	}

	// Fifth byte is command
	if packet[4] != responsePrepareOk {
		return nil, errInvalidPacketType
	}

	statementID := binary.LittleEndian.Uint32(packet[5:9])
	parametersNum := binary.LittleEndian.Uint16(packet[11:13])

	return &ComStmtPrepareOkResponse{StatementID: statementID, ParametersNum: parametersNum}, nil
}

// ComStmtExecuteRequest represents COM_STMT_EXECUTE request structure.
type ComStmtExecuteRequest struct {
	StatementID        uint32              // ID of prepared statement
	PreparedParameters []PreparedParameter // Slice of prepared parameters
}

// PreparedParameter structure represents single prepared parameter structure for COM_STMT_EXECUTE request.
type PreparedParameter struct {
	FieldType byte   // Type of prepared parameter. See https://mariadb.com/kb/en/mariadb/resultset/#field-types
	Flag      byte   // Unused
	Value     string // String value of any prepared parameter passed with COM_STMT_EXECUTE request
}

// DecodeComStmtExecuteRequest decodes COM_STMT_EXECUTE packet sent by MySQL client.
// Basic packet structure shown below.
// See https://mariadb.com/kb/en/mariadb/com_stmt_execute/
//
// int<3> PacketLength
// int<1> PacketNumber
// int<1> COM_STMT_EXECUTE (0x17)
// int<4> StatementID
// int<1> Flags
// int<4> IterationCount = 1
// if (ParamCount > 0)
// {
// 		byte<(ParamCount + 7) / 8> NullBitmap
// 		byte<1>: SendTypeToServer = 0 or 1
// 		if (SendTypeToServer)
//		{
// 			Foreach parameter
//			{
// 				byte<1>: FieldType
//				byte<1>: ParameterFlag
//			}
//		}
// 		Foreach parameter
//		{
// 			byte<n> BinaryParameterValue
//		}
// }
func DecodeComStmtExecuteRequest(packet []byte, paramsCount uint16) (*ComStmtExecuteRequest, error) {

	// Min packet length = header(4 bytes) + command(1 byte) + statementID(4 bytes)
	// + flags(1 byte) + iteration count(4 bytes)
	if err := CheckPacketLength(14, packet); err != nil {
		return nil, err
	}

	// Fifth byte is command
	if packet[4] != ComStmtExecute {
		return nil, errInvalidPacketType
	}

	r := bytes.NewReader(packet)

	// Skip packet header
	if err := SkipPacketHeader(r); err != nil {
		return nil, err
	}

	// Skip to statementID position
	if _, err := r.Seek(1, io.SeekCurrent); err != nil {
		return nil, err
	}

	// Read StatementID
	statementIDBuf := make([]byte, 4)
	if _, err := r.Read(statementIDBuf); err != nil {
		return nil, err
	}
	statementID := binary.LittleEndian.Uint32(statementIDBuf)

	// Skip to NullBitmap position
	if _, err := r.Seek(5, io.SeekCurrent); err != nil {
		return nil, err
	}

	// Make buffer for n=paramsCount prepared parameters
	parameters := make([]PreparedParameter, paramsCount)

	if paramsCount > 0 {
		nullBitmapLength := int64((paramsCount + 7) / 8)

		// Skip to SendTypeToServer position
		if _, err := r.Seek(nullBitmapLength, io.SeekCurrent); err != nil {
			return nil, err
		}

		// Read SendTypeToServer
		sendTypeToServer, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		if sendTypeToServer == 1 {
			for index := range parameters {

				// Read parameter FieldType and ParameterFlag
				parameterMeta := make([]byte, 2)
				if _, err := r.Read(parameterMeta); err != nil {
					return nil, err
				}

				parameters[index].FieldType = parameterMeta[0]
				parameters[index].Flag = parameterMeta[1]
			}
		}

		var fieldDecoderError error
		var fieldValue string

		for index, parameter := range parameters {
			switch parameter.FieldType {

			// MYSQL_TYPE_VAR_STRING (length encoded string)
			case fieldTypeString:
				fieldValue, fieldDecoderError = DecodeFieldTypeString(r)

			// MYSQL_TYPE_LONGLONG
			case fieldTypeLongLong:
				fieldValue, fieldDecoderError = DecodeFieldTypeLongLong(r)

			// MYSQL_TYPE_DOUBLE
			case fieldTypeDouble:
				fieldValue, fieldDecoderError = DecodeFieldTypeDouble(r)

			// Field with missing decoder
			default:
				return nil, errFieldTypeNotImplementedYet
			}

			// Return with first decoding error
			if fieldDecoderError != nil {
				return nil, fieldDecoderError
			}

			parameters[index].Value = fieldValue
			fieldValue = ""
		}
	}

	return &ComStmtExecuteRequest{StatementID: statementID, PreparedParameters: parameters}, nil
}

// DecodeFieldTypeString decodes MYSQL_TYPE_VAR_STRING field (length-encoded string)
// See https://mariadb.com/kb/en/mariadb/resultset/#field-types
func DecodeFieldTypeString(r *bytes.Reader) (string, error) {
	str, _, err := ReadLenEncodedString(r)

	// io.EOF is ok since reader may be empty already because of empty prepared parameter value
	if err != nil && err != io.EOF {
		return "", err
	}

	return str, nil
}

// DecodeFieldTypeLongLong decodes MYSQL_TYPE_LONGLONG field
// See https://mariadb.com/kb/en/mariadb/resultset/#field-types
func DecodeFieldTypeLongLong(r *bytes.Reader) (string, error) {
	var bigIntValue int64

	if err := binary.Read(r, binary.LittleEndian, &bigIntValue); err != nil {
		return "", nil
	}

	return strconv.FormatInt(bigIntValue, 10), nil
}

// DecodeFieldTypeDouble decodes MYSQL_TYPE_DOUBLE field
// See https://mariadb.com/kb/en/mariadb/resultset/#field-types
func DecodeFieldTypeDouble(r *bytes.Reader) (string, error) {
	// Read 8 bytes required for float64
	doubleLengthBuf := make([]byte, 8)
	if _, err := r.Read(doubleLengthBuf); err != nil {
		return "", err
	}

	doubleBits := binary.LittleEndian.Uint64(doubleLengthBuf)
	doubleValue := math.Float64frombits(doubleBits)

	return strconv.FormatFloat(doubleValue, 'f', doubleDecodePrecision, 64), nil
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
func ReadLenEncodedInteger(r *bytes.Reader) (value uint64, offset uint64) {
	firstLenEncIntByte, err := r.ReadByte()
	if err != nil {
		return 0, 0
	}

	switch firstLenEncIntByte {
	case 0xfb:
		value = 0
		offset = 1

	case 0xfc:
		data := make([]byte, 2)
		_, err = r.Read(data)
		if err != nil {
			return 0, 0
		}
		value = uint64(data[0]) | uint64(data[1])<<8
		offset = 3

	case 0xfd:
		data := make([]byte, 3)
		_, err = r.Read(data)
		if err != nil {
			return 0, 0
		}
		value = uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16
		offset = 4

	case 0xfe:
		data := make([]byte, 8)
		_, err = r.Read(data)
		if err != nil {
			return 0, 0
		}
		value = uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16 |
			uint64(data[3])<<24 | uint64(data[4])<<32 | uint64(data[5])<<40 |
			uint64(data[6])<<48 | uint64(data[7])<<56
		offset = 9

	default:
		value = uint64(firstLenEncIntByte)
		offset = 1
	}

	return value, offset
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
		} else {
			str = append(str, b)
		}
	}
}

// SkipPacketHeader rewinds reader to packet payload
func SkipPacketHeader(r *bytes.Reader) error {
	if _, err := r.Seek(4, io.SeekStart); err != nil {
		return err
	}

	return nil
}

// CheckPacketLength checks if packet length meets expected value
func CheckPacketLength(expected int, packet []byte) error {
	if len(packet) < expected {
		return errInvalidPacketLength
	}

	return nil
}
