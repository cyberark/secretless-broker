package provider

type Provider interface {
  Name() string
  Value(id string) ([]byte, error)
}
