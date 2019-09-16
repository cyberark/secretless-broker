package internal

// ProxyService is the interface the Secretless framework uses to start and stop
// ProxyServices. This interface allows ProxyServices to be implemented
// independent of each other. While operational, ProxyServices are expected to
// serve connections by routing them to the appropriate service connector
type ProxyService interface {
	// Start is a synchronous method responsible for carrying out the steps
	// necessary for a ProxyService to become ready to service client
	// connections.
	Start() error
	// Stop is a synchronous method responsible for terminating a ProxyService.
	// Stop is expected to both terminate the ProxyService and carry out any
	// necessary clean-up of resources held and consumed by the ProxyService
	// while operating.
	Stop() error
}

// CredentialsRetriever is a function signature for retrieval of credentials.
// The purpose of a CredentialsRetriever is to deliver credentials from within ProxyService instances and so it takes no arguments.
type CredentialsRetriever func() (map[string][]byte, error)
