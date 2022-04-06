package mock

import (
	"fmt"
	go_plugin "plugin"
)

// InvalidPlugin implements the rawPlugin interface
type InvalidPlugin struct {
	PluginAPIVersion string
	PluginType       string
	PluginID         string
	ErrorOnLookup    bool
}

// Lookup returns a go_plugin.Symbol for a given symbol name string. A go_plugin.Symbol
// is an empty interface, and can be instantiated with a function or a struct.
// This method is intended to mimic the Lookup method for a standard Go plugin.
func (r InvalidPlugin) Lookup(symbolName string) (go_plugin.Symbol, error) {
	if r.ErrorOnLookup {
		return nil, fmt.Errorf("error on lookup")
	}

	switch symbolName {
	case "PluginInfo":
		// Return something other than the correct type
		return func() string { return "test" }, nil
	}
	return nil, fmt.Errorf("unknown symbolName %s", symbolName)
}
