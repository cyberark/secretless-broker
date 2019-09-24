package internal

// CredentialsRetriever is a function signature for retrieval of credentials.
// The purpose of a CredentialsRetriever is to deliver credentials from within
// ProxyService instances and so it takes no arguments.
type CredentialsRetriever func() (map[string][]byte, error)

// Service is a generic service that can be started and stopped. We're currently
// using it to represent both the profile service and proxy services.
// TODO: The wisdom of an abstraction for a service that can be stopped/started
//   is something we want to revisit.  Standard functional command objects might
//   a better alternative, among other things.  We should revisit where we're
//   putting interfaces from a first pinciples/best practices perspective,
//   and create some policy around that.  For now, though, these aren't big
//   problems.
type Service interface {
	Start() error
	Stop() error
}
