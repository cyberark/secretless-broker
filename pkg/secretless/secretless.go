package secretless

// Service is a generic service that can be started and stopped.
type Service interface {
	Start() error
	Stop() error
}
