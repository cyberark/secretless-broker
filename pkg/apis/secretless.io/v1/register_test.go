package v1

import (
	"testing"

	secretlessio "github.com/cyberark/secretless-broker/pkg/apis/secretless.io"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSchemeGroupVersion(t *testing.T) {
	if SchemeGroupVersion.Group != secretlessio.GroupName {
		t.Fatalf("expected Group %q, got %q", secretlessio.GroupName, SchemeGroupVersion.Group)
	}
	if SchemeGroupVersion.Version != "v1" {
		t.Fatalf("expected Version 'v1', got %q", SchemeGroupVersion.Version)
	}
}

func TestResource(t *testing.T) {
	res := Resource("myres")
	if res.Group != SchemeGroupVersion.Group {
		t.Fatalf("expected Group %q, got %q", SchemeGroupVersion.Group, res.Group)
	}
	if res.Resource != "myres" {
		t.Fatalf("expected Resource 'myres', got %q", res.Resource)
	}
}

func TestAddKnownTypesRegistersTypes(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := addKnownTypes(scheme); err != nil {
		t.Fatalf("addKnownTypes returned error: %v", err)
	}
	kt := scheme.KnownTypes(SchemeGroupVersion)
	if _, ok := kt["Configuration"]; !ok {
		t.Fatalf("Configuration not registered in scheme for %v", SchemeGroupVersion)
	}
	if _, ok := kt["ConfigurationList"]; !ok {
		t.Fatalf("ConfigurationList not registered in scheme for %v", SchemeGroupVersion)
	}
}

func TestAddToSchemeBuilder(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme returned error: %v", err)
	}
	kt := scheme.KnownTypes(SchemeGroupVersion)
	if _, ok := kt["Configuration"]; !ok {
		t.Fatalf("Configuration not registered by AddToScheme for %v", SchemeGroupVersion)
	}
	if _, ok := kt["ConfigurationList"]; !ok {
		t.Fatalf("ConfigurationList not registered by AddToScheme for %v", SchemeGroupVersion)
	}
}
