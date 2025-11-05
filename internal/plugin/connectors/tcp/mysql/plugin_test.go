package mysql

import (
	"testing"

	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
)

func TestPluginInfo(t *testing.T) {
	info := PluginInfo()

	expected := map[string]string{
		"pluginAPIVersion": "0.1.0",
		"type":             "connector.tcp",
		"id":               "mysql",
		"description":      "returns an authenticated connection to a MySQL database",
	}

	if len(info) != len(expected) {
		t.Fatalf("unexpected PluginInfo length: got %d, want %d", len(info), len(expected))
	}

	for k, v := range expected {
		if got, ok := info[k]; !ok {
			t.Fatalf("missing key %q in PluginInfo", k)
		} else if got != v {
			t.Fatalf("unexpected PluginInfo[%q]: got %q, want %q", k, got, v)
		}
	}
}

func TestGetTCPPlugin_ReturnsConnectorConstructor(t *testing.T) {
	p := GetTCPPlugin()
	if p == nil {
		t.Fatal("GetTCPPlugin() returned nil")
	}

	cc, ok := p.(tcp.ConnectorConstructor)
	if !ok {
		t.Fatalf("GetTCPPlugin() is not tcp.ConnectorConstructor")
	}

	var res connector.Resources
	conn := cc(res)
	if conn == nil {
		t.Fatal("ConnectorConstructor returned nil connector")
	}

	if _, ok := conn.(tcp.ConnectorFunc); !ok {
		t.Fatalf("returned connector is not tcp.ConnectorFunc")
	}
}

func TestNewConnector_ReturnsConnectorFunc(t *testing.T) {
	var res connector.Resources
	conn := NewConnector(res)
	if conn == nil {
		t.Fatal("NewConnector returned nil")
	}
	if _, ok := conn.(tcp.ConnectorFunc); !ok {
		t.Fatalf("NewConnector did not return tcp.ConnectorFunc")
	}
}
