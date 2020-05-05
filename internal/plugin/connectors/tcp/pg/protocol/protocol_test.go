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
