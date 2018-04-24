package protocol

// ParseStartupMessage parses the pg startup message.
func ParseStartupMessage(message []byte) (version int32, options map[string]string, err error) {
	messageBuffer := NewMessageBuffer(message)
	if version, err = messageBuffer.ReadInt32(); err != nil {
		return
	}

	options = make(map[string]string)
	for {
		param, err := messageBuffer.ReadString()
		value, err := messageBuffer.ReadString()
		if err != nil || param == "\x00" {
			break
		}

		options[param] = value
	}

	return
}

// CreateStartupMessage creates a PG startup message. This message is used to
// startup all connections with a PG backend.
func CreateStartupMessage(username string, database string, options map[string]string) []byte {
	return []byte{}
}
