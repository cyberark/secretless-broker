package crd

import (
	"fmt"
	"testing"
	"time"

	api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitForConfig(t *testing.T, ch <-chan config_v2.Config, timeout time.Duration) config_v2.Config {
	select {
	case c := <-ch:
		return c
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for config")
	}
	panic("unreachable")
}

func Test_configurationManager_CRDDeleted_SendsEmptyConfig_Unit(t *testing.T) {
	m := &configurationManager{
		ConfigChangedChan: make(chan config_v2.Config, 1),
	}

	m.CRDDeleted(&api_v1.Configuration{})

	got := waitForConfig(t, m.ConfigChangedChan, 500*time.Millisecond)
	var want config_v2.Config
	if got.String() != want.String() {
		t.Fatalf("expected empty config from CRDDeleted; got: %v", got)
	}
}

func Test_configurationManager_ChannelSendReceive_Unit(t *testing.T) {
	m := &configurationManager{
		ConfigChangedChan: make(chan config_v2.Config, 1),
	}

	expected := config_v2.Config{}
	m.ConfigChangedChan <- expected

	got := waitForConfig(t, m.ConfigChangedChan, 500*time.Millisecond)
	if got.String() != expected.String() {
		t.Fatalf("channel roundtrip failed: got %v want %v", got, expected)
	}
}

func TestNewConfigChannel_Success_Unit(t *testing.T) {
	origInject := InjectCRD
	origRegister := RegisterCRDListener
	defer func() {
		InjectCRD = origInject
		RegisterCRDListener = origRegister
	}()

	InjectCRD = func() error { return nil }

	RegisterCRDListener = func(namespace string, configSpec string, reh ResourceEventHandler) error {
		if namespace != meta_v1.NamespaceAll {
			t.Fatalf("unexpected namespace: got %q want %q", namespace, meta_v1.NamespaceAll)
		}
		if configSpec != "my-spec" {
			t.Fatalf("configSpec not forwarded: got %q want %q", configSpec, "my-spec")
		}

		go func() {
			time.Sleep(10 * time.Millisecond)
			c := config_v2.Config{}
			cm := reh.(*configurationManager)
			cm.ConfigChangedChan <- c
		}()
		return nil
	}

	ch, err := NewConfigChannel("my-spec")
	if err != nil {
		t.Fatalf("NewConfigChannel returned unexpected error: %v", err)
	}

	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for config")
	case <-ch:
		// success
	}
}

func TestNewConfigChannel_InjectCRDError_Unit(t *testing.T) {
	origInject := InjectCRD
	origRegister := RegisterCRDListener
	defer func() {
		InjectCRD = origInject
		RegisterCRDListener = origRegister
	}()

	InjectCRD = func() error { return fmt.Errorf("inject failed") }

	RegisterCRDListener = func(namespace string, configSpec string, reh ResourceEventHandler) error {
		t.Fatalf("RegisterCRDListener should not be called when InjectCRD fails")
		return nil
	}

	ch, err := NewConfigChannel("any-spec")
	if err == nil {
		if ch != nil {
			select {
			case <-ch:
			case <-time.After(10 * time.Millisecond):
			}
		}
		t.Fatalf("expected error from NewConfigChannel when InjectCRD fails")
	}
}
