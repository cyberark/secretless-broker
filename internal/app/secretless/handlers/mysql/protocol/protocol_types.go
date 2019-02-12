package protocol

import (
	"bytes"
	"fmt"
)

type MySQLInt struct {
	length int
	value  int
}

func NewMySQLInt(reader *bytes.Reader, len int) (*MySQLInt, error) {
	val := make([]byte, len)
	if _, err := reader.Read(val); err != nil {
		return nil, err
	}

	if len > 4 {
		return nil, fmt.Errorf("NewMySQLInt len must be <= to 4")
	}

	value := uint32(val[0])
	for i := 1; i < len; i++  {
		value |= uint32(val[i])<< uint(i * 8)
	}
	return &MySQLInt{
		length: len,
		value:  int(value),
	}, nil
}

func (my *MySQLInt) Val() int { // Go representation
	return my.value
}

func (my *MySQLInt) Bytes() []byte { // Go representation
	data := make([]byte, my.length)

	for i := 0; i < my.length; i++  {
		data[i] = byte(my.value >> uint(i * 8))
	}
	return data
}

func (my *MySQLInt) Pack(buff *bytes.Buffer) error  { // Raw representation
	_, err := buff.Write(my.Bytes())
	return err
}

// MySQLNString
type MySQLNString struct {
	length  int
	value  string
}

func NewMySQLNString(reader *bytes.Reader, len int) (*MySQLNString, error) {
	val := make([]byte, len)
	if _, err := reader.Read(val); err != nil {
		return nil, err
	}

	return &MySQLNString{
		length: len,
		value:  string(val),
	}, nil
}

func (my *MySQLNString) Bytes() []byte { // Go representation
	return []byte([]byte(my.value[:my.length]))
}

func (my *MySQLNString) Val() string { // Go representation
	return my.value
}

func (my *MySQLNString) Pack(buff *bytes.Buffer) error  { // Raw representation
	_, err := buff.Write([]byte(my.value[:my.length]))
	return err
}

// MySQLString
type MySQLString struct {
	value  string
}

func NewMySQLString(reader *bytes.Reader) (*MySQLString, error) {
	var data []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == 0x00 {
			break
		}
		data = append(data, b)
	}

	return &MySQLString{
		value:  string(data),
	}, nil
}

func (my *MySQLString) Val() string { // Go representation
	return my.value
}

func (my *MySQLString) Bytes() []byte { // Go representation
	return []byte(my.value)
}

func (my *MySQLString) Pack(buff *bytes.Buffer) error  { // Raw representation
	if _, err := buff.WriteString(my.value); err != nil {
		return err
	}

	return buff.WriteByte(0)
}