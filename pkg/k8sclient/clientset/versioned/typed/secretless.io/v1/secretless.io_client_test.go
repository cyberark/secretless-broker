package v1_test

import (
	"testing"

	"github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	client "github.com/cyberark/secretless-broker/pkg/k8sclient/clientset/versioned/typed/secretless.io/v1"
	"k8s.io/client-go/rest"
)

func TestNewForConfig_ReturnsErrorOnInvalidConfig(t *testing.T) {
	cfg := &rest.Config{Host: "://bad-url"}
	cli, err := client.NewForConfig(cfg)
	if err == nil {
		t.Errorf("Expected error on invalid config, got nil")
	}
	if cli != nil {
		t.Errorf("Expected nil client on error, got non-nil")
	}
}

func TestNewForConfig_ReturnsClient(t *testing.T) {
	cfg := &rest.Config{
		Host:    "http://localhost",
		APIPath: "/apis",
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: nil,
		},
	}
	cli, err := client.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("NewForConfig returned error: %v", err)
	}
	if cli == nil {
		t.Fatalf("NewForConfig returned nil client")
	}
}

func TestNew_ReturnsClient(t *testing.T) {
	var restClient rest.Interface // nil is fine for this test
	cli := client.New(restClient)
	if cli == nil {
		t.Fatalf("New returned nil client")
	}
}

func TestRESTClient_ReturnsRestClient(t *testing.T) {
	var restClient rest.Interface // nil is fine for this test
	cli := client.New(restClient)
	if cli.RESTClient() != restClient {
		t.Errorf("RESTClient did not return the expected rest.Interface")
	}
}

func TestConfigurations_ReturnsNonNil(t *testing.T) {
	var restClient rest.Interface // nil is fine for this test
	cli := client.New(restClient)
	c := cli.Configurations("default")
	if c == nil {
		t.Errorf("Configurations returned nil")
	}
}
