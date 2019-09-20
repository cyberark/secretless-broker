package internal

// CredentialsRetriever is a function signature for retrieval of credentials.
// The purpose of a CredentialsRetriever is to deliver credentials from within
// ProxyService instances and so it takes no arguments.
type CredentialsRetriever func() (map[string][]byte, error)
