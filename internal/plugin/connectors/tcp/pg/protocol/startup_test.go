package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStartupMessage_Success(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("user")
	msg.WriteString("testuser")
	msg.WriteString("database")
	msg.WriteString("testdb")
	msg.WriteString("application_name")
	msg.WriteString("myapp")
	msg.WriteByte(0x00)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "testuser", options["user"])
	assert.Equal(t, "testdb", options["database"])
	assert.Equal(t, "myapp", options["application_name"])
}

func TestParseStartupMessage_SSLRequest(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(SSLRequestCode)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, SSLRequestCode, version)
	assert.Empty(t, options)
}

func TestParseStartupMessage_MinimalOptions(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("user")
	msg.WriteString("testuser")
	msg.WriteByte(0x00)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "testuser", options["user"])
	assert.Equal(t, 1, len(options))
}

func TestParseStartupMessage_MultipleOptions(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("user")
	msg.WriteString("testuser")
	msg.WriteString("database")
	msg.WriteString("testdb")
	msg.WriteString("client_encoding")
	msg.WriteString("UTF8")
	msg.WriteString("DateStyle")
	msg.WriteString("ISO, MDY")
	msg.WriteString("TimeZone")
	msg.WriteString("UTC")
	msg.WriteByte(0x00)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, 5, len(options))
	assert.Equal(t, "testuser", options["user"])
	assert.Equal(t, "testdb", options["database"])
	assert.Equal(t, "UTF8", options["client_encoding"])
	assert.Equal(t, "ISO, MDY", options["DateStyle"])
	assert.Equal(t, "UTC", options["TimeZone"])
}

func TestParseStartupMessage_EmptyMessage(t *testing.T) {
	version, options, err := ParseStartupMessage([]byte{})

	assert.Error(t, err)
	assert.Equal(t, int32(0), version)
	assert.Nil(t, options)
}

func TestParseStartupMessage_OnlyVersion(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteByte(0x00)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Empty(t, options)
}

func TestParseStartupMessage_OddNumberOfParameters(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("user")
	msg.WriteString("testuser")
	msg.WriteString("database")
	msg.WriteByte(0x00)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "testuser", options["user"])
}

func TestParseStartupMessage_DuplicateKeys(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("user")
	msg.WriteString("firstuser")
	msg.WriteString("user")
	msg.WriteString("seconduser")
	msg.WriteByte(0x00)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "seconduser", options["user"])
}

func TestParseStartupMessage_EmptyValues(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("user")
	msg.WriteString("")
	msg.WriteString("database")
	msg.WriteString("")
	msg.WriteByte(0x00)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "", options["user"])
	assert.Equal(t, "", options["database"])
}

func TestParseStartupMessage_EmptyKeys(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("")
	msg.WriteString("value")
	msg.WriteString("normalkey")
	msg.WriteString("normalvalue")
	msg.WriteByte(0x00)

	version, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "value", options[""])
	assert.Equal(t, "normalvalue", options["normalkey"])
}

func TestParseStartupMessage_SpecialCharactersInValues(t *testing.T) {
	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("user")
	msg.WriteString("user@host.com")
	msg.WriteString("application_name")
	msg.WriteString("My Application (v1.0)")
	msg.WriteByte(0x00)

	_, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, "user@host.com", options["user"])
	assert.Equal(t, "My Application (v1.0)", options["application_name"])
}

func TestParseStartupMessage_LongValues(t *testing.T) {
	longValue := string(make([]byte, 1000))
	for i := range longValue {
		longValue = longValue[:i] + "x" + longValue[i+1:]
	}

	msg := NewMessageBuffer([]byte{})
	msg.WriteInt32(ProtocolVersion)
	msg.WriteString("user")
	msg.WriteString(longValue)
	msg.WriteByte(0x00)

	_, options, err := ParseStartupMessage(msg.Bytes())

	require.NoError(t, err)
	assert.Equal(t, longValue, options["user"])
}

func TestCreateStartupMessage_Basic(t *testing.T) {
	options := map[string]string{
		"client_encoding": "UTF8",
		"DateStyle":       "ISO, MDY",
	}

	result := CreateStartupMessage(ProtocolVersion, "testuser", "testdb", options)

	version, parsedOptions, err := ParseStartupMessage(result[4:])

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "testuser", parsedOptions["user"])
	assert.Equal(t, "testdb", parsedOptions["database"])
	assert.Equal(t, "UTF8", parsedOptions["client_encoding"])
	assert.Equal(t, "ISO, MDY", parsedOptions["DateStyle"])
}

func TestCreateStartupMessage_NoOptions(t *testing.T) {
	result := CreateStartupMessage(ProtocolVersion, "testuser", "testdb", nil)

	version, parsedOptions, err := ParseStartupMessage(result[4:])

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "testuser", parsedOptions["user"])
	assert.Equal(t, "testdb", parsedOptions["database"])
	assert.Equal(t, 2, len(parsedOptions))
}

func TestCreateStartupMessage_EmptyOptions(t *testing.T) {
	options := map[string]string{}

	result := CreateStartupMessage(ProtocolVersion, "testuser", "testdb", options)

	_, parsedOptions, err := ParseStartupMessage(result[4:])

	require.NoError(t, err)
	assert.Equal(t, "testuser", parsedOptions["user"])
	assert.Equal(t, "testdb", parsedOptions["database"])
}

func TestCreateStartupMessage_MessageStructure(t *testing.T) {
	result := CreateStartupMessage(ProtocolVersion, "user", "db", nil)

	assert.True(t, len(result) >= 4)
	assert.Equal(t, byte(0x00), result[len(result)-1])

	msg := NewMessageBuffer(result[4:])
	ver, readErr := msg.ReadInt32()
	require.NoError(t, readErr)
	assert.Equal(t, ProtocolVersion, ver)
}

func TestCreateStartupMessage_RoundTrip(t *testing.T) {
	originalOptions := map[string]string{
		"client_encoding":  "UTF8",
		"application_name": "test_app",
		"TimeZone":         "UTC",
	}

	created := CreateStartupMessage(ProtocolVersion, "myuser", "mydb", originalOptions)
	version, parsedOptions, err := ParseStartupMessage(created[4:])

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "myuser", parsedOptions["user"])
	assert.Equal(t, "mydb", parsedOptions["database"])

	for key, value := range originalOptions {
		assert.Equal(t, value, parsedOptions[key])
	}
}

func TestCreateStartupMessage_EmptyUsernameAndDatabase(t *testing.T) {
	result := CreateStartupMessage(ProtocolVersion, "", "", nil)

	version, parsedOptions, err := ParseStartupMessage(result[4:])

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, "", parsedOptions["user"])
	assert.Equal(t, "", parsedOptions["database"])
}

func TestCreateStartupMessage_SpecialCharacters(t *testing.T) {
	options := map[string]string{
		"special": "value with spaces and @#$%",
	}

	result := CreateStartupMessage(ProtocolVersion, "user@domain", "db-name", options)
	_, parsedOptions, err := ParseStartupMessage(result[4:])

	require.NoError(t, err)
	assert.Equal(t, "user@domain", parsedOptions["user"])
	assert.Equal(t, "db-name", parsedOptions["database"])
	assert.Equal(t, "value with spaces and @#$%", parsedOptions["special"])
}

func TestCreateStartupMessage_ManyOptions(t *testing.T) {
	options := map[string]string{
		"option1":  "value1",
		"option2":  "value2",
		"option3":  "value3",
		"option4":  "value4",
		"option5":  "value5",
		"option6":  "value6",
		"option7":  "value7",
		"option8":  "value8",
		"option9":  "value9",
		"option10": "value10",
	}

	result := CreateStartupMessage(ProtocolVersion, "user", "db", options)
	version, parsedOptions, err := ParseStartupMessage(result[4:])

	require.NoError(t, err)
	assert.Equal(t, ProtocolVersion, version)
	assert.Equal(t, 12, len(parsedOptions))

	for key, value := range options {
		assert.Equal(t, value, parsedOptions[key])
	}
}
