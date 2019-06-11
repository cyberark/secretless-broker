package api

type DbTesterOptions struct {
	ConnectionType string
	DatabaseName   string
	Debug          bool
	Host           string
	Password       string
	Port           string
	SslMode        string
	Socket         string
	Username       string
}

type DbTester interface {
	Connect(DbTesterOptions) error
	GetQueryMarkers(length int) string
	Query(string, ...interface{}) error
	QueryRows(string, string, ...interface{}) ([]string, error)
	Shutdown() error
}
