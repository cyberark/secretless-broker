package testutil

// ClientConfiguration specifies the username, password, and SSL setting
// for test cases.
type ClientConfiguration struct {
	SSL      bool
	Username string
	Password string
}
