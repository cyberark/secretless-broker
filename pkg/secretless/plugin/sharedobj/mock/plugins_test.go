package mock

import (
	"reflect"
	"testing"
)

func TestHTTPAndTCPInternalExternalLists(t *testing.T) {
	intHTTP := HTTPInternalPluginsByID()
	extHTTP := HTTPExternalPluginsByID()
	if len(intHTTP) != 3 {
		t.Fatalf("expected 3 internal HTTP plugins, got %d", len(intHTTP))
	}
	if len(extHTTP) != 2 {
		t.Fatalf("expected 2 external HTTP plugins, got %d", len(extHTTP))
	}

	intTCP := TCPInternalPluginsByID()
	extTCP := TCPExternalPluginsByID()
	if len(intTCP) != 3 {
		t.Fatalf("expected 3 internal TCP plugins, got %d", len(intTCP))
	}
	if len(extTCP) != 3 {
		t.Fatalf("expected 3 external TCP plugins, got %d", len(extTCP))
	}
}

func TestAllHTTPAndAllTCPMerge(t *testing.T) {
	intHTTP := HTTPInternalPluginsByID()
	extHTTP := HTTPExternalPluginsByID()
	allHTTP := AllHTTPPlugins()
	if len(allHTTP) != len(intHTTP)+len(extHTTP) {
		t.Fatalf("AllHTTPPlugins length mismatch: got %d, want %d", len(allHTTP), len(intHTTP)+len(extHTTP))
	}

	intTCP := TCPInternalPluginsByID()
	extTCP := TCPExternalPluginsByID()
	allTCP := AllTCPPlugins()
	if len(allTCP) != len(intTCP)+len(extTCP) {
		t.Fatalf("AllTCPPlugins length mismatch: got %d, want %d", len(allTCP), len(intTCP)+len(extTCP))
	}
}

func TestInternalAndExternalPluginsHelpers(t *testing.T) {
	ip, err := GetInternalPlugins()
	if err != nil {
		t.Fatalf("GetInternalPlugins returned error: %v", err)
	}
	if len(ip.HTTPPlugins()) == 0 && len(ip.TCPPlugins()) == 0 {
		t.Fatalf("GetInternalPlugins returned empty plugin sets")
	}

	ep, err := GetExternalPlugins("", "", NewLogger())
	if err != nil {
		t.Fatalf("GetExternalPlugins returned error: %v", err)
	}
	if len(ep.HTTPPlugins()) == 0 && len(ep.TCPPlugins()) == 0 {
		t.Fatalf("GetExternalPlugins returned empty plugin sets")
	}
}

func TestNewLoggerNotNil(t *testing.T) {
	if NewLogger() == nil {
		t.Fatalf("NewLogger returned nil")
	}
}

func TestRawPluginLookupAndSymbols(t *testing.T) {
	isNilValue := func(v reflect.Value) bool {
		switch v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			return v.IsNil()
		default:
			return false
		}
	}

	for key, rp := range RawPlugins {
		sym, err := rp.Lookup("PluginInfo")
		if err != nil {
			t.Fatalf("Lookup PluginInfo for %s returned error: %v", key, err)
		}
		infoFunc, ok := sym.(func() map[string]string)
		if !ok {
			t.Fatalf("PluginInfo symbol has wrong type for %s", key)
		}
		info := infoFunc()
		if info["pluginAPIVersion"] != pluginAPIVersion {
			t.Fatalf("unexpected pluginAPIVersion for %s: got %s want %s", key, info["pluginAPIVersion"], pluginAPIVersion)
		}
		if info["id"] == "" || info["type"] == "" {
			t.Fatalf("plugin info missing fields for %s: %+v", key, info)
		}

		if info["type"] == "connector.http" {
			symHTTP, err := rp.Lookup("GetHTTPPlugin")
			if err != nil {
				t.Fatalf("Lookup GetHTTPPlugin for %s returned error: %v", key, err)
			}

			if hf, ok := symHTTP.(func() interface{}); ok {
				if hf() == nil {
					t.Fatalf("GetHTTPPlugin returned nil for %s", key)
				}
			} else {
				v := reflect.ValueOf(symHTTP)
				if v.Kind() != reflect.Func {
					t.Fatalf("GetHTTPPlugin symbol has wrong type for %s", key)
				}
				res := v.Call(nil)
				if len(res) == 0 || isNilValue(res[0]) {
					t.Fatalf("GetHTTPPlugin returned nil for %s", key)
				}
			}
		}

		if info["type"] == "connector.tcp" {
			symTCP, err := rp.Lookup("GetTCPPlugin")
			if err != nil {
				t.Fatalf("Lookup GetTCPPlugin for %s returned error: %v", key, err)
			}

			if tf, ok := symTCP.(func() interface{}); ok {
				if tf() == nil {
					t.Fatalf("GetTCPPlugin returned nil for %s", key)
				}
			} else {
				v := reflect.ValueOf(symTCP)
				if v.Kind() != reflect.Func {
					t.Fatalf("GetTCPPlugin symbol has wrong type for %s", key)
				}
				res := v.Call(nil)
				if len(res) == 0 || isNilValue(res[0]) {
					t.Fatalf("GetTCPPlugin returned nil for %s", key)
				}
			}
		}
	}

	_, err := RawPlugins["http1"].Lookup("UnknownSymbol")
	if err == nil {
		t.Fatalf("expected error for unknown symbol, got nil")
	}
}
