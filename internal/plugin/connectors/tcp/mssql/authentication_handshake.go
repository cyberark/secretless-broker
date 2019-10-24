package mssql

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"sort"

	mssql "github.com/denisenkom/go-mssqldb"
)

// PerformFakeHandshakeForClient carries out a fake handshake with a client
// 1. reads the prelogin packet
// 2. writes the prelogin repsonse declaring NO support for encryption
// 3. reads the login packet
// 4. writes the login response as a success. TODO: what about failure ?
func PerformFakeHandshakeForClient(clientConn net.Conn) error {
	// using the default packet size of 4096 (see go-mssqldb/conn_str.go)
	clientBuf := mssql.NewTdsBuffer(4096, clientConn)
	msg, err := ReadClientPrelogin(clientBuf)
	if err != nil {
		return fmt.Errorf("bad prelogin: %s", err)
	}

	// based on wireshark, the prelogin response from the server is almost identical
	// to the login message from the client, with minor changes as below:
	msg[mssql.PreloginTHREADID] = []byte{}
	msg[mssql.PreloginVERSION] = []byte{0x0e, 0x00, 0x0c, 0xa6, 0x00, 0x00}
	msg[mssql.PreloginENCRYPTION] = []byte{mssql.EncryptNotSup}

	err = WritePreloginResponse(clientBuf, msg)
	if err != nil {
		return fmt.Errorf("bad prelogin response: %s", err)
	}

	// we actually don't care what the client has to say.
	// we just need for them to not be blocked
	err = clientBuf.ReadNextPacket()
	if err != nil {
		return fmt.Errorf("bad client login read: %s", err)
	}

	// created an unencrypted connection to mssql, and just copied the raw bytes
	// this has loginack, and done tokens.
	_, err = clientConn.Write([]byte{
		0x04, 0x01, 0x00, 0x64, 0x00, 0x37, 0x01, 0x00, 0xad, 0x36, 0x00, 0x01, 0x74, 0x00, 0x00, 0x04, 0x16, 0x4d, 0x00, 0x69, 0x00, 0x63, 0x00, 0x72, 0x00, 0x6f, 0x00, 0x73, 0x00, 0x6f, 0x00, 0x66, 0x00, 0x74, 0x00, 0x20, 0x00, 0x53, 0x00, 0x51, 0x00, 0x4c, 0x00, 0x20, 0x00, 0x53, 0x00, 0x65, 0x00, 0x72, 0x00, 0x76, 0x00, 0x65, 0x00, 0x72, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0e, 0x00, 0x0c, 0xa6, 0xe3, 0x13, 0x00, 0x04, 0x04, 0x34, 0x00, 0x30, 0x00, 0x39, 0x00, 0x36, 0x00, 0x04, 0x34, 0x00, 0x30, 0x00, 0x39, 0x00, 0x36, 0x00, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})

	if err != nil {
		return fmt.Errorf("bad client write login response: %s", err)
	}

	return nil
}

// reads the client prelogin
func ReadClientPrelogin(r *mssql.TdsBuffer) (map[uint8][]byte, error) {
	_, err := r.BeginRead()
	if err != nil {
		return nil, err
	}

	struct_buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	offset := 0
	results := map[uint8][]byte{}
	for true {
		rec_type := struct_buf[offset]
		if rec_type == mssql.PreloginTERMINATOR {
			break
		}

		rec_offset := binary.BigEndian.Uint16(struct_buf[offset+1:])
		rec_len := binary.BigEndian.Uint16(struct_buf[offset+3:])
		value := struct_buf[rec_offset : rec_offset+rec_len]
		results[rec_type] = value
		offset += 5
	}
	return results, nil
}

// writes the prelogin response as though from the mssql backend
func WritePreloginResponse(w *mssql.TdsBuffer, fields map[uint8][]byte) error {
	var err error

	w.BeginPacket(4, false)
	offset := uint16(5*len(fields) + 1)
	keys := make(mssql.KeySlice, 0, len(fields))
	for k, _ := range fields {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	// writing header
	for _, k := range keys {
		err = w.WriteByte(k)
		if err != nil {
			return err
		}
		err = binary.Write(w, binary.BigEndian, offset)
		if err != nil {
			return err
		}
		v := fields[k]
		size := uint16(len(v))
		err = binary.Write(w, binary.BigEndian, size)
		if err != nil {
			return err
		}
		offset += size
	}
	err = w.WriteByte(mssql.PreloginTERMINATOR)
	if err != nil {
		return err
	}
	// writing values
	for _, k := range keys {
		v := fields[k]
		written, err := w.Write(v)
		if err != nil {
			return err
		}
		if written != len(v) {
			return errors.New("Write method didn't write the whole value")
		}
	}
	return w.FinishPacket()
}
