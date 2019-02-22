package mysql

import "strconv"

// ConnectionDetails stores the connection info to the real backend database.
// These values are pulled from the handler credentials config
type ConnectionDetails struct {
	Host     string
	Port     uint
	Username string
	Password string
	Options  map[string]string
}

// Address returns a string representing the network location (host and port)
// of a MySQL server.  This is the string you would would typically use to
// connect to the database -- eg, in cmd line tools.
//
func (cd *ConnectionDetails) Address() string {
	return cd.Host + ":" + strconv.FormatUint(uint64(cd.Port), 10)
}
