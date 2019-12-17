package mssql

// CreateAuthenticationOKMessage creates an MSSQL message which indicates
// successful authentication.
func (connector *SingleUseConnector) CreateAuthenticationOKMessage() []byte {
	// The packet format - https://www.freetds.org/tds.html#packet
	// The login ack response - https://www.freetds.org/tds.html#t173

	// TODO: This message should stay static but it's better to build it the same way
	// we do in the pg connector -
	// https://github.com/cyberark/secretless-broker/blob/master/internal/plugin/connectors/tcp/pg/protocol/auth.go#L150

	// TODO: Also check if we can (and need to) use the actual TDS version & server
	//  details using go-mssqldb.
	//  If so, we need to merge this PR - https://github.com/cyberark/go-mssqldb/pull/3

	// Create a hard-coded OK response message
	message := []byte{
		// header
		// [ REPLY packet type, last packet indicator, packet size (2 bytes),
		// channel (2 bytes, can be zeroed), packet number, window (can be zeroed) ]
		0x04, 0x01, 0x00, 0x4E, 0x00, 0x00, 0x01, 0x00,

		// LoginAck Token
		// Login ack indicator
		0xad,
		// Packet length - 54 bytes from the next one until the Done Token (not including)
		0x36, 0x00,
		// ack success - TODO: test with TDS 5.0, because there the success ack is 0x05
		0x01,
		// TDS version - TODO: verify that it can be hard coded regardless to
		//  the actual TDS version of the server
		0x74, 0x00, 0x00, 0x04,
		// server name length - 22 chars
		0x16,
		// server name - 'Microsoft SQL Server'
		0x4d, 0x00, 0x69, 0x00, 0x63, 0x00, 0x72, 0x00, 0x6f, 0x00, 0x73, 0x00,
		0x6f, 0x00, 0x66, 0x00, 0x74, 0x00, 0x20, 0x00, 0x53, 0x00, 0x51, 0x00,
		0x4c, 0x00, 0x20, 0x00, 0x53, 0x00, 0x65, 0x00, 0x72, 0x00, 0x76, 0x00,
		0x65, 0x00, 0x72, 0x00, 0x00, 0x00, 0x00, 0x00,
		// server version - TODO: verify that it can be hard coded regardless
		//  to the actual server version
		0x0e, 0x00, 0x0c, 0xa6,

		// Done Token - indicates the end of the packet
		0xfd, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	return message
}

// CreateGenericErrorMessage creates an MSSQL error message
func (connector *SingleUseConnector) CreateGenericErrorMessage() []byte {
	// The packet format - https://www.freetds.org/tds.html#packet
	// The Error token - https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-tds/9805e9fa-1f8b-4cf8-8f78-8d2602228635

	// TODO: Create the error object with go-mssqldb. In order to use it, we should
	// add there a function to convert the object to a byte array

	// Create a hard-coded error response message
	message := []byte{
		// header
		// [ REPLY packet type, last packet indicator, packet size (2 bytes),
		// channel (2 bytes, can be zeroed), packet number, window (can be zeroed) ]
		0x04, 0x01, 0x00, 0x7a, 0x00, 0x33, 0x01, 0x00,

		// Error Token
		// Error token indicator
		0xaa,
		// Token length - 2 bytes
		0x62, 0x00,
		// SQL Error Number - currently using 18456 (login failed for user)
		// TODO: Find generic error number
		0x18, 0x48, 0x00, 0x00,
		// state - TODO: better understand this.
		0x01,
		// severity - 16 indicates a general error that can be corrected by the user.
		0x0e,
		// Error message length
		0x1e, 0x00,
		// Error message: "Generic SQL Error"
		0x47, 0x00, 0x65, 0x00, 0x6e, 0x00, 0x65, 0x00, 0x72, 0x00,
		0x69, 0x00, 0x63, 0x00, 0x20, 0x00, 0x53, 0x00, 0x51, 0x00,
		0x4c, 0x00, 0x20, 0x00, 0x45, 0x00, 0x72, 0x00, 0x72, 0x00,
		0x6f, 0x00, 0x72, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// server name length - TODO: Change this according to the name
		0x0c,
		// server name - TODO: Understand this and change accordingly.
		//  This is the value received from the sql server in the test
		0x61, 0x00, 0x64, 0x00, 0x30, 0x00, 0x39, 0x00,
		0x37, 0x00, 0x33, 0x00, 0x31, 0x00, 0x37, 0x00,
		0x35, 0x00, 0x38, 0x00, 0x33, 0x00, 0x35, 0x00,
		// process name length (can be zero)
		0x00,
		// Line number - zero indicates that it's not related to an SQL batch line
		0x00, 0x00, 0x00, 0x00,
		// Done Token - indicates the end of the packet
		0xfd,
		0x02, 0x00,
		0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	return message
}
