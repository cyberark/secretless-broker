package client

// Config is abstract and represents the configurations that apply to all
// clients, each RunQuery translates the configuration into a form that makes
// sense for its client.
// e.g. Username, Database translate to the following command for sqlcmd:
//
// sqlcmd -d Database -U Username
//
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	// This is in relation to what is generally referred to as Application Intent.
	// It can only take 2 values, ReadWrite or ReadOnly.
	ReadOnly bool
}

// RunQuery represents the invocation of an MSSQL client. It takes two arguments,
// database client configuration (Config) and query (string). It returns a string
// and an error; the string captures the success output and the error captures the failure.
//
// As an example, SqlcmdExec is of type RunQuery. SqlcmdExec invokes the sqlcmd
// program using the arguments provided. An example invocation might look as follows:
//
// sqlcmd -d Database -U Username -Q query
//
type RunQuery func(cfg Config, query string) (string, error)

// ConcurrentCall calls RunQuery concurrently, and returns a channel
// that can be waited on to get the client response.
func (runQuery RunQuery) ConcurrentCall(cfg Config, query string) chan Response {
	resChan := make(chan Response)

	go func() {
		out, err := runQuery(
			cfg,
			query,
		)

		resChan <- Response{
			Out: out,
			Err: err,
		}
	}()

	return resChan
}

// Response represents the response from calling a RunQuery. It is composed
// of a string and an error; the string captures the success output and the error
// captures the failure.
type Response struct {
	Out string
	Err error
}
