package protocol

import (
	"encoding/binary"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadMessage(t *testing.T) {
	t.Run("parses contents", func(t *testing.T) {
		r, w := net.Pipe()
		expectedMessageType := byte(12)
		expectedMessage := []byte{0, 1, 2, 3, 4}

		go func() {
			err := binary.Write(w, binary.BigEndian, expectedMessageType)
			if err != nil {
				panic(err)
			}
			err = binary.Write(w, binary.BigEndian, int32(len(expectedMessage)+4))
			if err != nil {
				panic(err)
			}

			_, err = w.Write(expectedMessage)
			if err != nil {
				panic(err)
			}
		}()
		messageType, message, err := ReadMessage(r)

		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, expectedMessage, message)
		assert.Equal(t, expectedMessageType, messageType)
	})

	t.Run("validates message length", func(t *testing.T) {
		r, w := net.Pipe()
		expectedMessageType := byte(12)
		// a message length less than 4 is invalid
		expectedMessageLength := int32(3)

		go func() {
			err := binary.Write(w, binary.BigEndian, expectedMessageType)
			if err != nil {
				panic(err)
			}
			err = binary.Write(w, binary.BigEndian, expectedMessageLength)
			if err != nil {
				panic(err)
			}
		}()
		_, _, err := ReadMessage(r)

		if !assert.Error(t, err) {
			return
		}
		assert.Contains(t, err.Error(), "invalid message length")
	})
}

func TestReadStartupMessage(t *testing.T) {
	t.Run("parses startup message successfully", func(t *testing.T) {
		r, w := net.Pipe()
		expectedMessage := []byte{0, 1, 2, 3, 4}

		go func() {
			err := binary.Write(w, binary.BigEndian, int32(len(expectedMessage)+4))
			if err != nil {
				panic(err)
			}

			_, err = w.Write(expectedMessage)
			if err != nil {
				panic(err)
			}
		}()

		message, err := ReadStartupMessage(r)

		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, expectedMessage, message)
	})

	t.Run("validates message length", func(t *testing.T) {
		r, w := net.Pipe()
		invalidMessageLength := int32(3)

		go func() {
			err := binary.Write(w, binary.BigEndian, invalidMessageLength)
			if err != nil {
				panic(err)
			}
		}()

		_, err := ReadStartupMessage(r)

		if !assert.Error(t, err) {
			return
		}
		assert.Contains(t, err.Error(), "invalid message length")
	})

	t.Run("handles read errors", func(t *testing.T) {
		r, w := net.Pipe()

		go func() {
			err := binary.Write(w, binary.BigEndian, int32(10))
			if err != nil {
				panic(err)
			}
			w.Close()
		}()

		_, err := ReadStartupMessage(r)

		assert.Error(t, err)
	})
}

func TestReadMessage_Errors(t *testing.T) {
	t.Run("handles error reading message type", func(t *testing.T) {
		r, w := net.Pipe()
		w.Close()

		_, _, err := ReadMessage(r)

		assert.Error(t, err)
	})

	t.Run("handles error reading message body", func(t *testing.T) {
		r, w := net.Pipe()

		go func() {
			err := binary.Write(w, binary.BigEndian, byte(12))
			if err != nil {
				panic(err)
			}
			err = binary.Write(w, binary.BigEndian, int32(10))
			if err != nil {
				panic(err)
			}
			w.Close()
		}()

		_, _, err := ReadMessage(r)

		assert.Error(t, err)
	})
}
