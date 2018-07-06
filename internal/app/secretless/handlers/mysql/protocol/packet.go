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
	"io"
	"net"
)

// ConnSettings contains the connection settings
type ConnSettings struct {
	ClientCapabilities uint32
	ServerCapabilities uint32
	SelectedDb         string
}

// DeprecateEOFSet checks whether ClientDeprecateEOF is set on the client or server
func (h *ConnSettings) DeprecateEOFSet() bool {
	return ((ClientDeprecateEOF & h.ServerCapabilities) != 0) &&
		((ClientDeprecateEOF & h.ClientCapabilities) != 0)
}

// ProcessHandshake handles handshake between server and client.
// Returns server and client handshake responses
func ProcessHandshake(client net.Conn, mysql net.Conn) (*HandshakeV10, *HandshakeResponse41, error) {

	// Read server handshake
	packet, err := ProxyPacket(mysql, client)
	if err != nil {
		println(err.Error())
		return nil, nil, err
	}

	serverHandshake, err := UnpackHandshakeV10(packet)
	if err != nil {
		println(err.Error())
		return nil, nil, err
	}

	// Read client handshake response
	packet, err = ProxyPacket(client, mysql)
	if err != nil {
		println(err.Error())
		return nil, nil, err
	}

	clientHandshake, err := UnpackHandshakeResponse41(packet)
	if err != nil {
		println(err.Error())
		return nil, nil, err
	}

	// Read server OK response
	if _, err = ProxyPacket(mysql, client); err != nil {
		println(err.Error())
		return nil, nil, err
	}

	return serverHandshake, clientHandshake, nil
}

// ReadPrepareResponse reads response from MySQL server for COM_STMT_PREPARE
// query issued by client.
// ...
func ReadPrepareResponse(conn net.Conn) ([]byte, byte, error) {
	pkt, err := ReadPacket(conn)
	if err != nil {
		return nil, 0, err
	}

	switch pkt[4] {
	case responsePrepareOk:
		numParams := binary.LittleEndian.Uint16(pkt[9:11])
		numColumns := binary.LittleEndian.Uint16(pkt[11:13])
		packetsExpected := 0

		if numParams > 0 {
			packetsExpected += int(numParams) + 1
		}

		if numColumns > 0 {
			packetsExpected += int(numColumns) + 1
		}

		var data []byte
		var eofCnt int

		data = append(data, pkt...)

		for i := 1; i <= packetsExpected; i++ {
			eofCnt++
			pkt, err = ReadPacket(conn)
			if err != nil {
				return nil, 0, err
			}

			data = append(data, pkt...)
		}

		return data, responseOk, nil

	case ResponseErr:
		return pkt, ResponseErr, nil
	}

	return nil, 0, nil
}

// ReadErrMessage reads the message in an error packet
func ReadErrMessage(errPacket []byte) string {
	return string(errPacket[13:])
}

// ReadShowFieldsResponse reads the response with deprecateEof set to true
func ReadShowFieldsResponse(conn net.Conn) ([]byte, byte, error) {
	return ReadResponse(conn, true)
}

// ReadResponse reads the response
func ReadResponse(conn net.Conn, deprecateEOF bool) ([]byte, byte, error) {
	pkt, err := ReadPacket(conn)
	if err != nil {
		return nil, 0, err
	}

	switch pkt[4] {
	case responseOk:
		return pkt, responseOk, nil

	case ResponseErr:
		return pkt, ResponseErr, nil

	case responseLocalinfile:
	}

	var data []byte

	data = append(data, pkt...)

	if !deprecateEOF {
		pktReader := bytes.NewReader(pkt[4:])
		columns, _ := ReadLenEncodedInteger(pktReader)

		toRead := int(columns) + 1

		for i := 0; i < toRead; i++ {
			pkt, err := ReadPacket(conn)
			if err != nil {
				return nil, 0, err
			}

			data = append(data, pkt...)
		}
	}

	for {
		pkt, err := ReadPacket(conn)
		if err != nil {
			return nil, 0, err
		}

		data = append(data, pkt...)

		if pkt[4] == responseEOF {
			break
		}
	}

	return data, responseResultset, nil
}

// ReadPacket ...
func ReadPacket(conn net.Conn) ([]byte, error) {

	// Read packet header
	header := []byte{0, 0, 0, 0}
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	// Calculate packet body length
	bodyLen := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	// Read packet body
	body := make([]byte, bodyLen)
	n, err := io.ReadFull(conn, body)
	if err != nil {
		return nil, err
	}

	return append(header, body[0:n]...), nil
}

// WritePacket ...
func WritePacket(pkt []byte, conn net.Conn) (int, error) {
	n, err := conn.Write(pkt)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// ProxyPacket ...
func ProxyPacket(src, dst net.Conn) ([]byte, error) {
	pkt, err := ReadPacket(src)
	if err != nil {
		return nil, err
	}

	_, err = WritePacket(pkt, dst)
	if err != nil {
		return nil, err
	}

	return pkt, nil
}
