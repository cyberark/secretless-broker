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
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnpackOkResponse(t *testing.T) {

	type UnpackOkResponseAssert struct {
		Packet   []byte
		HasError bool
		Error    error
		OkResponse
	}

	testData := []*UnpackOkResponseAssert{
		{
			[]byte{
				0x30, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x22, 0x00, 0x00, 0x00, 0x28, 0x52, 0x6f, 0x77, 0x73,
				0x20, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x64, 0x3a, 0x20, 0x31, 0x20, 0x20, 0x43, 0x68, 0x61,
				0x6e, 0x67, 0x65, 0x64, 0x3a, 0x20, 0x31, 0x20, 0x20, 0x57, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67,
				0x73, 0x3a, 0x20, 0x30,
			},
			false,
			nil,
			OkResponse{
				PacketType:   0x00,
				AffectedRows: uint64(1),
				LastInsertID: uint64(0),
				StatusFlags:  uint16(34),
				Warnings:     uint16(0)},
		},
		{
			[]byte{0x07, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00},
			false,
			nil,
			OkResponse{
				PacketType:   0x00,
				AffectedRows: uint64(0),
				LastInsertID: uint64(0),
				StatusFlags:  uint16(2),
				Warnings:     uint16(0)},
		},
		{
			[]byte{0x07, 0x00, 0x00, 0x01, 0x00, 0x01, 0x02, 0x02, 0x00, 0x00, 0x00},
			false,
			nil,
			OkResponse{
				PacketType:   0x00,
				AffectedRows: uint64(1),
				LastInsertID: uint64(2),
				StatusFlags:  uint16(2),
				Warnings:     uint16(0)},
		},
	}

	for _, asserted := range testData {
		decoded, err := UnpackOkResponse(asserted.Packet)

		assert.Nil(t, err)

		if err == nil {
			assert.Equal(t, asserted.OkResponse.PacketType, decoded.PacketType)
			assert.Equal(t, asserted.OkResponse.AffectedRows, decoded.AffectedRows)
			assert.Equal(t, asserted.OkResponse.LastInsertID, decoded.LastInsertID)
			assert.Equal(t, asserted.OkResponse.StatusFlags, decoded.StatusFlags)
			assert.Equal(t, asserted.OkResponse.Warnings, decoded.Warnings)
		}
	}
}

func TestUnpackHandshakeV10(t *testing.T) {

	type UnpackHandshakeV10Assert struct {
		Packet   []byte
		HasError bool
		Error    error
		HandshakeV10
		CapabilitiesMap map[uint32]bool
	}

	testData := []*UnpackHandshakeV10Assert{
		{
			[]byte{
				0x4a, 0x00, 0x00, 0x00, 0x0a, 0x35, 0x2e, 0x35, 0x2e, 0x35, 0x36, 0x00, 0x5e, 0x06, 0x00, 0x00,
				0x48, 0x6a, 0x5b, 0x6a, 0x24, 0x71, 0x30, 0x3a, 0x00, 0xff, 0xf7, 0x08, 0x02, 0x00, 0x0f, 0x80,
				0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6f, 0x43, 0x40, 0x56, 0x6e,
				0x4b, 0x68, 0x4a, 0x79, 0x46, 0x30, 0x5a, 0x00, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x6e, 0x61,
				0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x00,
			},
			false,
			nil,
			HandshakeV10{
				ProtocolVersion:    byte(10),
				ServerVersion:      "5.5.56",
				ConnectionID:       uint32(1630),
				AuthPlugin:         "mysql_native_password",
				ServerCapabilities: binary.LittleEndian.Uint32([]byte{255, 247, 15, 128}),
				Salt: []byte{0x48, 0x6a, 0x5b, 0x6a, 0x24, 0x71, 0x30, 0x3a, 0x6f, 0x43, 0x40, 0x56, 0x6e, 0x4b,
					0x68, 0x4a, 0x79, 0x46, 0x30, 0x5a},
			},
			map[uint32]bool{
				ClientLongPassword: true, ClientFoundRows: true, ClientLongFlag: true,
				ClientConnectWithDB: true, ClientNoSchema: true, ClientCompress: true, ClientODBC: true,
				ClientLocalFiles: true, ClientIgnoreSpace: true, ClientProtocol41: true, ClientInteractive: true,
				ClientSSL: false, ClientIgnoreSIGPIPE: true, ClientTransactions: true, ClientMultiStatements: true,
				ClientMultiResults: true, ClientPSMultiResults: true, ClientPluginAuth: true, ClientConnectAttrs: false,
				ClientPluginAuthLenEncClientData: false, ClientCanHandleExpiredPasswords: false,
				ClientSessionTrack: false, ClientDeprecateEOF: false},
		},
		{
			[]byte{
				0x4a, 0x00, 0x00, 0x00, 0x0a, 0x35, 0x2e, 0x37, 0x2e, 0x31, 0x38, 0x00, 0x0f, 0x00, 0x00, 0x00,
				0x15, 0x12, 0x4b, 0x1f, 0x70, 0x2b, 0x33, 0x55, 0x00, 0xff, 0xff, 0x08, 0x02, 0x00, 0xff, 0xc1,
				0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x30, 0x0d, 0x0a, 0x28,
				0x06, 0x4a, 0x12, 0x5e, 0x45, 0x18, 0x05, 0x00, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x6e, 0x61,
				0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x00,
			},
			false,
			nil,
			HandshakeV10{
				ProtocolVersion:    byte(10),
				ServerVersion:      "5.7.18",
				ConnectionID:       uint32(15),
				AuthPlugin:         "mysql_native_password",
				ServerCapabilities: binary.LittleEndian.Uint32([]byte{255, 255, 255, 193}),
				Salt: []byte{0x15, 0x12, 0x4b, 0x1f, 0x70, 0x2b, 0x33, 0x55, 0x01, 0x30, 0x0d,
					0x0a, 0x28, 0x06, 0x4a, 0x12, 0x5e, 0x45, 0x18, 0x05},
			},
			map[uint32]bool{
				ClientLongPassword: true, ClientFoundRows: true, ClientLongFlag: true,
				ClientConnectWithDB: true, ClientNoSchema: true, ClientCompress: true, ClientODBC: true,
				ClientLocalFiles: true, ClientIgnoreSpace: true, ClientProtocol41: true, ClientInteractive: true,
				ClientSSL: true, ClientIgnoreSIGPIPE: true, ClientTransactions: true, ClientMultiStatements: true,
				ClientMultiResults: true, ClientPSMultiResults: true, ClientPluginAuth: true, ClientConnectAttrs: true,
				ClientPluginAuthLenEncClientData: true, ClientCanHandleExpiredPasswords: true,
				ClientSessionTrack: true, ClientDeprecateEOF: true},
		},
	}

	for _, asserted := range testData {
		decoded, err := UnpackHandshakeV10(asserted.Packet)

		if err != nil {
			assert.Equal(t, asserted.Error, err)
		} else {
			assert.Equal(t, asserted.HandshakeV10.ProtocolVersion, decoded.ProtocolVersion)
			assert.Equal(t, asserted.HandshakeV10.ServerVersion, decoded.ServerVersion)
			assert.Equal(t, asserted.HandshakeV10.ConnectionID, decoded.ConnectionID)
			assert.Equal(t, asserted.HandshakeV10.AuthPlugin, decoded.AuthPlugin)
			assert.Equal(t, asserted.HandshakeV10.Salt, decoded.Salt)
			assert.Equal(t, asserted.HandshakeV10.ServerCapabilities, decoded.ServerCapabilities)

			for flag, isSet := range asserted.CapabilitiesMap {
				if isSet {
					assert.True(t, decoded.ServerCapabilities&flag > 0)
					if decoded.ServerCapabilities&flag == 0 {
						println(flag)
					}
				} else {
					assert.True(t, decoded.ServerCapabilities&flag == 0)
				}
			}
		}
	}
}

func TestPackHandshakeV10(t *testing.T) {
	input := &HandshakeV10{
		ProtocolVersion:    byte(10),
		ServerVersion:      "5.5.56",
		ConnectionID:       uint32(1630),
		AuthPlugin:         "mysql_native_password",
		ServerCapabilities: binary.LittleEndian.Uint32([]byte{255, 247, 15, 128}),
		Salt: []byte{0x48, 0x6a, 0x5b, 0x6a, 0x24, 0x71, 0x30, 0x3a, 0x6f, 0x43, 0x40, 0x56, 0x6e, 0x4b,
			0x68, 0x4a, 0x79, 0x46, 0x30, 0x5a},
	}

	output, err := PackHandshakeV10(input)
	newInput, err := UnpackHandshakeV10(output)

	assert.Equal(t, input, newInput)
	assert.Equal(t, nil, err)
}

func TestUnpackHandshakeResponse41(t *testing.T) {
	expected := HandshakeResponse41{
		SequenceID:      1,
		CapabilityFlags: uint32(33464965),
		MaxPacketSize:   uint32(1073741824),
		ClientCharset:   uint8(8),
		// file deepcode ignore NoHardcodedCredentials/test: This is a test file
		Username:       "roger",
		AuthLength:     int64(20),
		AuthPluginName: "mysql_native_password",
		AuthResponse: []byte{0xc0, 0xb, 0xbc, 0xb6, 0x6, 0xf5,
			0x4f, 0x4e, 0xf4, 0x1b, 0x87, 0xc0, 0xb8, 0x89, 0xae,
			0xc4, 0x49, 0x7c, 0x46, 0xf3},
		Database: "",
		PacketTail: []byte{0x58, 0x3, 0x5f, 0x6f, 0x73, 0xa, 0x6d, 0x61,
			0x63, 0x6f, 0x73, 0x31, 0x30, 0x2e, 0x31, 0x32, 0xc, 0x5f, 0x63,
			0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x8,
			0x6c, 0x69, 0x62, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x4, 0x5f, 0x70,
			0x69, 0x64, 0x5, 0x36, 0x36, 0x34, 0x37, 0x39, 0xf, 0x5f, 0x63,
			0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
			0x6f, 0x6e, 0x6, 0x35, 0x2e, 0x37, 0x2e, 0x32, 0x30, 0x9, 0x5f,
			0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x6, 0x78, 0x38,
			0x36, 0x5f, 0x36, 0x34},
	}
	input := []byte{0xaa, 0x0, 0x0, 0x1, 0x85, 0xa2, 0xfe, 0x1, 0x0,
		0x0, 0x0, 0x40, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x72, 0x6f, 0x67, 0x65, 0x72, 0x0, 0x14, 0xc0,
		0xb, 0xbc, 0xb6, 0x6, 0xf5, 0x4f, 0x4e, 0xf4, 0x1b, 0x87, 0xc0,
		0xb8, 0x89, 0xae, 0xc4, 0x49, 0x7c, 0x46, 0xf3, 0x6d, 0x79, 0x73,
		0x71, 0x6c, 0x5f, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x70,
		0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x0, 0x58, 0x3, 0x5f,
		0x6f, 0x73, 0xa, 0x6d, 0x61, 0x63, 0x6f, 0x73, 0x31, 0x30, 0x2e,
		0x31, 0x32, 0xc, 0x5f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f,
		0x6e, 0x61, 0x6d, 0x65, 0x8, 0x6c, 0x69, 0x62, 0x6d, 0x79, 0x73,
		0x71, 0x6c, 0x4, 0x5f, 0x70, 0x69, 0x64, 0x5, 0x36, 0x36, 0x34,
		0x37, 0x39, 0xf, 0x5f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f,
		0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x6, 0x35, 0x2e, 0x37,
		0x2e, 0x32, 0x30, 0x9, 0x5f, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f,
		0x72, 0x6d, 0x6, 0x78, 0x38, 0x36, 0x5f, 0x36, 0x34}

	output, err := UnpackHandshakeResponse41(input)

	assert.Equal(t, expected, *output)
	assert.Equal(t, nil, err)
}

func TestInjectCredentials(t *testing.T) {
	username := "testuser" // 8
	// file deepcode ignore HardcodedPassword/test: This is a test file
	password := "testpass"
	salt := []byte{0x2f, 0x50, 0x25, 0x34, 0x78, 0x17, 0x1, 0x44, 0x1d,
		0xc, 0x61, 0x4f, 0x5c, 0x69, 0x65, 0x6f, 0x25, 0x66, 0x7c, 0x64}
	expectedAuth := []byte{0xf, 0xf8, 0xe1, 0xa3, 0xe7, 0xe3, 0x5f, 0xd2,
		0xb1, 0x69, 0x8c, 0x39, 0x5b, 0xfa, 0x99, 0x4f, 0x53, 0xdd, 0xe5,
		0x35} // 20
	expectedHeader := []byte{0xa2, 0x0, 0x0, 0x1}
	// expectedHeader[0] = 0xaa + (8 - 14) + (20 - 20) + (21 - 23)

	// test with handshake response that already has auth set to another value
	handshake := HandshakeResponse41{
		AuthLength:     int64(20),
		AuthPluginName: "caching_sha256_password", // 23
		AuthResponse: []byte{0xc0, 0xb, 0xbc, 0xb6, 0x6, 0xf5, 0x4f, 0x4e,
			0xf4, 0x1b, 0x87, 0xc0, 0xb8, 0x89, 0xae, 0xc4, 0x49, 0x7c, 0x46, 0xf3}, // 20
		Username:   "madeupusername", // 14
		SequenceID: 1,
	}

	err := InjectCredentials("mysql_native_password", &handshake, salt, username, password)

	assert.Equal(t, username, handshake.Username)
	assert.Equal(t, int64(20), handshake.AuthLength)
	assert.Equal(t, expectedAuth, handshake.AuthResponse)
	assert.Equal(t, expectedHeader[3], handshake.SequenceID)
	assert.Equal(t, nil, err)

	// test with handshake response with empty auth and mysql_native_password
	expectedHeader = []byte{0xb8, 0x0, 0x0, 0x1}
	// expectedHeader[0] = 0xaa + (8 - 14) + (20 - 0) + (21 - 21)
	handshake = HandshakeResponse41{
		AuthLength:     0,
		AuthPluginName: "mysql_native_password", // 21
		AuthResponse:   []byte{},                // 0
		Username:       "madeupusername",        // 14
		SequenceID:     1,
	}

	err = InjectCredentials("mysql_native_password", &handshake, salt, username, password)

	assert.Equal(t, username, handshake.Username)
	assert.Equal(t, int64(20), handshake.AuthLength)
	assert.Equal(t, expectedAuth, handshake.AuthResponse)
	assert.Equal(t, expectedHeader[3], handshake.SequenceID)
	assert.Equal(t, nil, err)
}

func TestPackHandshakeResponse41(t *testing.T) {
	input := &HandshakeResponse41{
		SequenceID:      1,
		CapabilityFlags: uint32(33464965),
		MaxPacketSize:   uint32(1073741824),
		ClientCharset:   uint8(8),
		Username:        "roger",
		AuthLength:      int64(20),
		AuthResponse: []byte{0xc0, 0xb, 0xbc, 0xb6, 0x6, 0xf5,
			0x4f, 0x4e, 0xf4, 0x1b, 0x87, 0xc0, 0xb8, 0x89, 0xae,
			0xc4, 0x49, 0x7c, 0x46, 0xf3},
		Database: "",
		PacketTail: []byte{0x58, 0x3, 0x5f, 0x6f, 0x73, 0xa, 0x6d, 0x61,
			0x63, 0x6f, 0x73, 0x31, 0x30, 0x2e, 0x31, 0x32, 0xc, 0x5f, 0x63,
			0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x8,
			0x6c, 0x69, 0x62, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x4, 0x5f, 0x70,
			0x69, 0x64, 0x5, 0x36, 0x36, 0x34, 0x37, 0x39, 0xf, 0x5f, 0x63,
			0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
			0x6f, 0x6e, 0x6, 0x35, 0x2e, 0x37, 0x2e, 0x32, 0x30, 0x9, 0x5f,
			0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x6, 0x78, 0x38,
			0x36, 0x5f, 0x36, 0x34},
	}

	output, err := PackHandshakeResponse41(input)
	newInput, err := UnpackHandshakeResponse41(output)

	assert.Equal(t, input, newInput)
	assert.Equal(t, nil, err)
}

func TestGetLenEncodedIntegerSize(t *testing.T) {
	inputArray := []byte{0xfc, 0xfd, 0xfe, 0xfb}
	expectedArray := []byte{2, 3, 8, 1}

	for k, v := range inputArray {
		output := GetLenEncodedIntegerSize(v)

		assert.Equal(t, expectedArray[k], output)
	}
}

func TestReadLenEncodedInteger(t *testing.T) {
	expected := uint64(251)
	input := bytes.NewReader([]byte{0xfc, 0xfb, 0x00})

	output, err := ReadLenEncodedInteger(input)

	assert.Equal(t, expected, output)
	assert.Equal(t, nil, err)
}

func TestReadLenEncodedString(t *testing.T) {
	expected := "ABCDEFGHIKLMONPQRSTYW"
	packet := bytes.NewReader([]byte{
		0x15, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4b, 0x4c, 0x4d, 0x4f, 0x4e, 0x50,
		0x51, 0x52, 0x53, 0x54, 0x59, 0x57})

	decoded, length, err := ReadLenEncodedString(packet)

	assert.Equal(t, expected, decoded)
	assert.Equal(t, len(expected), int(length))
	assert.Equal(t, nil, err)
}

func TestReadEOFLengthString(t *testing.T) {
	expected := "SET sql_mode='STRICT_TRANS_TABLES'"
	encoded := []byte{
		0x53, 0x45, 0x54, 0x20, 0x73, 0x71, 0x6c, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x3d, 0x27, 0x53, 0x54, 0x52,
		0x49, 0x43, 0x54, 0x5f, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x5f, 0x54, 0x41, 0x42, 0x4c, 0x45, 0x53, 0x27,
	}

	decoded := ReadEOFLengthString(encoded)

	assert.Equal(t, expected, decoded)
}

func TestReadNullTerminatedString(t *testing.T) {
	x := bytes.NewReader([]byte{0x35, 0x2e, 0x37, 0x2e, 0x31, 0x38, 0x00})
	assert.Equal(t, "5.7.18", ReadNullTerminatedString(x))
}

func TestReadNullTerminatedBytes(t *testing.T) {
	input := bytes.NewReader([]byte{0x1d, 0xc, 0x61, 0x4f, 0x5c, 0x69,
		0x65, 0x6f, 0x25, 0x66, 0x7c, 0x64, 0x0, 0x6d, 0x79, 0x73, 0x71,
		0x6c, 0x5f, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61,
		0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x0})
	expected := []byte{0x1d, 0xc, 0x61, 0x4f, 0x5c, 0x69, 0x65, 0x6f,
		0x25, 0x66, 0x7c, 0x64}

	output := ReadNullTerminatedBytes(input)

	assert.Equal(t, expected, output)
}

func TestGetPacketHeader(t *testing.T) {
	input := bytes.NewReader([]byte{0x4a, 0x0, 0x0, 0x0, 0xa, 0x35, 0x2e,
		0x37, 0x2e, 0x32, 0x31, 0x0, 0x38, 0x9, 0x0, 0x0, 0x2f, 0x50, 0x25,
		0x34, 0x78, 0x17, 0x1, 0x44, 0x0, 0xff, 0xff, 0x8, 0x2, 0x0,
		0xff, 0xc1, 0x15, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x1d, 0xc, 0x61, 0x4f, 0x5c, 0x69, 0x65, 0x6f, 0x25, 0x66,
		0x7c, 0x64, 0x0, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x6e, 0x61,
		0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f,
		0x72, 0x64, 0x0})
	expected := []byte{0x4a, 0x0, 0x0, 0x0}

	output, err := GetPacketHeader(input)

	assert.Equal(t, expected, output)
	assert.Equal(t, nil, err)
}

func TestCheckPacketLength(t *testing.T) {
	inputLength := 4
	inputPacket := []byte{0xf, 0xf8, 0xe1, 0xa3}

	err := CheckPacketLength(inputLength, inputPacket)

	assert.Equal(t, nil, err)
}

func TestNativePassword(t *testing.T) {
	expected := []byte{0xf, 0xf8, 0xe1, 0xa3, 0xe7, 0xe3, 0x5f, 0xd2, 0xb1, 0x69,
		0x8c, 0x39, 0x5b, 0xfa, 0x99, 0x4f, 0x53, 0xdd, 0xe5, 0x35}
	inputPassword := "testpass"
	inputSalt := []byte{0x2f, 0x50, 0x25, 0x34, 0x78, 0x17, 0x1, 0x44,
		0x1d, 0xc, 0x61, 0x4f, 0x5c, 0x69, 0x65, 0x6f, 0x25, 0x66, 0x7c, 0x64}

	output, err := NativePassword([]byte(inputPassword), inputSalt)

	assert.Equal(t, expected, output)
	assert.Equal(t, nil, err)
}

func TestCreateAuthResponse(t *testing.T) {
	testCases := []struct {
		authPlugin  string
		password    []byte
		salt        []byte
		expectedLen int
		expectErr   bool
	}{
		{
			authPlugin:  "mysql_native_password",
			password:    []byte("password"),
			salt:        []byte("salt"),
			expectedLen: 20,
		},
		{
			authPlugin:  "caching_sha2_password",
			password:    []byte("password"),
			salt:        []byte("salt"),
			expectedLen: 32,
		},
		{
			authPlugin: "unknown_auth_plugin",
			password:   []byte("password"),
			salt:       []byte("salt"),
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		actual, err := CreateAuthResponse(tc.authPlugin, tc.password, tc.salt)
		if tc.expectErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedLen, len(actual))
		}
	}
}

func TestUnpackAuthSwitchRequest(t *testing.T) {
	testCases := []struct {
		name          string
		input         []byte
		expectedError string
		expectedReq   *AuthSwitchRequest
	}{
		{
			name: "valid AuthSwitchRequest packet",
			input: []byte{
				0x02, 0x00, 0x00, 0x01, // Header
				0x01, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x00, // Plugin name ("plugin")
				0x01, 0x02, 0x03, // Plugin data
			},
			expectedReq: &AuthSwitchRequest{
				SequenceNumber: 1,
				PluginName:     "plugin",
				PluginData:     []byte{0x01, 0x02, 0x03},
			},
		},
		{
			name:          "missing plugin name",
			input:         []byte{0x00, 0x00, 0x00, 0x01},
			expectedError: "Invalid AuthSwitchRequest packet: Missing plugin name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := UnpackAuthSwitchRequest(tc.input)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, req)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, tc.expectedReq.SequenceNumber, req.SequenceNumber)
				assert.Equal(t, tc.expectedReq.PluginName, req.PluginName)
				assert.Equal(t, tc.expectedReq.PluginData, req.PluginData)
			}
		})
	}
}

func TestUnpackAuthRequestPubKeyResponse(t *testing.T) {
	// Generate a test RSA public key
	testPubKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	testPubKeyBytes, _ := x509.MarshalPKIXPublicKey(&testPubKey.PublicKey)
	testPubKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: testPubKeyBytes,
	})
	testPubKeyBytes = append([]byte{0x02, 0x00, 0x00, 0x04, 0x01}, testPubKeyPem...)

	testCases := []struct {
		name          string
		input         []byte
		expectedError string
		expectedResp  *rsa.PublicKey
	}{
		{
			name:         "valid AuthRequestPubKeyResponse packet",
			input:        testPubKeyBytes,
			expectedResp: &testPubKey.PublicKey,
		},
		{
			name:          "missing RSA public key",
			input:         []byte{0x02, 0x00, 0x00, 0x04, 0x01},
			expectedError: "no pem data found, data: ",
		},
		{
			name:          "invalid RSA public key",
			input:         []byte{0x02, 0x00, 0x00, 0x04, 0x01, 0x01, 0x02},
			expectedError: "no pem data found, data: \x01\x02",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := UnpackAuthRequestPubKeyResponse(tc.input)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.EqualValues(t, tc.expectedResp, resp)
			}
		})
	}
}

func TestPackAuthSwitchResponse(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	seqID := uint8(9)

	expected := []byte{
		0x03, 0x00, 0x00, 0x09, // Header (including sequence number)
		0x01, 0x02, 0x03, // Data
	}

	output, err := PackAuthSwitchResponse(seqID, data)
	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

func TestUnpackAuthMoreDataResponse(t *testing.T) {
	testCases := []struct {
		name          string
		input         []byte
		expectedError string
		expectedResp  *AuthMoreDataResponse
	}{
		{
			name:  "valid AuthMoreDataResponse packet",
			input: []byte{0x01, 0x00, 0x00, 0x09, 0x01, 0x04},
			expectedResp: &AuthMoreDataResponse{
				SequenceID: 9,
				PacketType: 1,
				StatusTag:  4,
			},
		},
		{
			name:          "missing data",
			input:         []byte{0x01, 0x00, 0x00, 0x09},
			expectedError: ErrInvalidPacketLength.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := UnpackAuthMoreDataResponse(tc.input)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}
