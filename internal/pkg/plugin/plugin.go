package plugin

import (
	"net"
	"plugin"

	"github.com/conjurinc/secretless/pkg/secretless"
)

type Plugin struct {
	_funcInitialize      plugin.Symbol
	_funcCreateListener  plugin.Symbol
	_funcNewConnection   plugin.Symbol
	_funcCloseConnection plugin.Symbol
	_funcCreateHandler   plugin.Symbol
	_funcDestroyHandler  plugin.Symbol
	_funcResolveVariable plugin.Symbol
	_funcClientData      plugin.Symbol
	_funcServerData      plugin.Symbol
}

func (p *Plugin) Initialize() {
	p._funcInitialize.(func())()
}
func (p *Plugin) CreateListener(l secretless.Listener) {
	p._funcCreateListener.(func(secretless.Listener))(l)
}
func (p *Plugin) NewConnection(c net.Conn) {
	p._funcNewConnection.(func(net.Conn))(c)
}
func (p *Plugin) CloseConnection(c net.Conn) {
	p._funcCloseConnection.(func(net.Conn))(c)
}
func (p *Plugin) CreateHandler(l secretless.Listener, h secretless.Handler) {
	p._funcCreateHandler.(func(secretless.Listener, secretless.Handler))(l, h)
}
func (p *Plugin) DestroyHandler(h secretless.Handler) {
	p._funcDestroyHandler.(func(secretless.Handler))(h)
}
func (p *Plugin) ResolveVariable(provider secretless.Provider, id string, value []byte) {
	p._funcResolveVariable.(func(secretless.Provider, string, []byte))(provider, id, value)
}
func (p *Plugin) ClientData(buf []byte) {
	p._funcClientData.(func([]byte))(buf)
}
func (p *Plugin) ServerData(buf []byte) {
	p._funcServerData.(func([]byte))(buf)
}
