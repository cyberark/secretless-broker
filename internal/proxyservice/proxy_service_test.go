package proxyservice

import (
	"net"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"errors"
	"net/http"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/cyberark/secretless-broker/internal"
	v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
	logapi "github.com/cyberark/secretless-broker/pkg/secretless/log"
	"github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector"
	httpplugin "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/http"
	tcpplugin "github.com/cyberark/secretless-broker/pkg/secretless/plugin/connector/tcp"
	"github.com/stretchr/testify/assert"
)

// --- Unit Tests ---

func TestZeroizeCredentials(t *testing.T) {
	creds := map[string][]byte{
		"foo": []byte("test1"),
		"baz": []byte("test2"),
	}
	internal.ZeroizeCredentials(creds)
	for _, v := range creds {
		for _, b := range v {
			assert.Equal(t, byte(0), b)
		}
	}
}

func TestProxyServicesStart(t *testing.T) {
	ps := &proxyServices{
		availPlugins:  &fakeAvailablePlugins{},
		configsByType: v2.ConfigsByType{TCP: []v2.Service{{Name: "svc", Connector: "mock", ListenOn: v2.NetworkAddress("tcp://localhost:0")}}},
		logger:        &testLogger{},
		resolver:      &mockResolver{},
	}

	err := ps.Start()
	assert.NoError(t, err)
	assert.Len(t, ps.runningServices, 1)
}

func TestProxyServicesStop(t *testing.T) {
	ms1, ms2 := &mockService{}, &mockService{}
	ps := &proxyServices{logger: &testLogger{}, runningServices: []internal.Service{ms1, ms2}}

	assert.NoError(t, ps.Stop())
	assert.True(t, ms1.stopped)
	assert.True(t, ms2.stopped)

	ms1.stopped, ms2.stopped = false, false
	ms2.stopErr = assert.AnError

	ps.runningServices = []internal.Service{ms1, ms2}
	assert.Error(t, ps.Stop())
	assert.Contains(t, ps.Stop().Error(), assert.AnError.Error())
	assert.True(t, ms1.stopped)
	assert.True(t, ms2.stopped)
}

func TestProxyServicesCredsRetriever(t *testing.T) {
	mock := &mockResolver{}
	ps := &proxyServices{resolver: mock}
	retriever := ps.credsRetriever(nil)
	creds, err := retriever()
	assert.NoError(t, err)
	assert.Equal(t, []byte("bar"), creds["foo"])
	assert.True(t, mock.called, "resolver should have been called")
}

func TestHandleErrors_ExitsOnError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Exit test not supported on Windows")
	}
	if os.Getenv("TEST_HANDLE_ERRORS_EXIT") == "1" {
		handleErrors(validation.Errors{"svc": errors.New("fail")}, &testLogger{})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestHandleErrors_ExitsOnError")
	cmd.Env = append(os.Environ(), "TEST_HANDLE_ERRORS_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process did not exit as expected, err=%v", err)
}

func TestCreateServiceFailures(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Exit test not supported on Windows")
	}
	if os.Getenv("TEST_FAIL_SERVICE_EXIT") != "" {
		failType := os.Getenv("TEST_FAIL_SERVICE_EXIT")
		ps := &proxyServices{
			logger:   &testLogger{},
			resolver: &mockResolver{},
		}
		switch failType {
		case "tcp":
			ps.configsByType = v2.ConfigsByType{TCP: []v2.Service{{Name: "svc", Connector: "mock", ListenOn: v2.NetworkAddress("tcp://bad:0")}}}
			ps.availPlugins = &fakeAvailablePlugins{}
		case "http":
			ps.configsByType = v2.ConfigsByType{HTTP: []v2.HTTPServiceConfig{{
				SharedListenOn: v2.NetworkAddress("tcp://bad:0"),
				SubserviceConfigs: []v2.Service{{Name: "svc", Connector: "mock", ConnectorConfig: []byte("{}")}},
			}}}
			ps.availPlugins = &fakeAvailablePlugins{}
		case "ssh":
			ps.configsByType = v2.ConfigsByType{SSH: []v2.Service{{Name: "svc", Connector: "mock", ListenOn: v2.NetworkAddress("tcp://bad:0")}}}
			ps.availPlugins = &fakeAvailablePlugins{}
		case "sshagent":
			ps.configsByType = v2.ConfigsByType{SSHAgent: []v2.Service{{Name: "svc", Connector: "mock", ListenOn: v2.NetworkAddress("tcp://bad:0")}}}
			ps.availPlugins = &fakeAvailablePlugins{}
		}
		ps.Start()
		return
	}
	types := []struct {
		name      string
		env       string
	}{
		{"TCP listener fail", "tcp"},
		{"HTTP listener fail", "http"},
		{"SSH listener fail", "ssh"},
		{"SSHAgent listener fail", "sshagent"},
	}
	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestCreateServiceFailures")
			cmd.Env = append(os.Environ(), "TEST_FAIL_SERVICE_EXIT="+tt.env)
			err := cmd.Run()
			if e, ok := err.(*exec.ExitError); ok && !e.Success() {
				return
			}
			t.Fatalf("process did not exit as expected for %s, err=%v", tt.name, err)
		})
	}
}

func TestCreateHTTPService_ConfigParseFail(t *testing.T) {
	ps := &proxyServices{
		logger:        &testLogger{},
		resolver:      &mockResolver{},
	}
	badCfg := v2.HTTPServiceConfig{
		SharedListenOn: v2.NetworkAddress("tcp://localhost:0"),
		SubserviceConfigs: []v2.Service{{Connector: "mock", ConnectorConfig: []byte("bad: [!!!]")}},
	}
	ps.availPlugins = &fakeAvailablePluginsWithHTTPMock{}
	_, err := ps.createHTTPService(badCfg, ps.availPlugins.HTTPPlugins())
	assert.Error(t, err)
}

func TestCreateSSHAndSSHAgentService_SuccessAndListenerFail(t *testing.T) {
	ps := &proxyServices{logger: &testLogger{}, resolver: &mockResolver{}}
	types := []struct {
		name     string
		createFn func(cfg v2.Service) (internal.Service, error)
	}{
		{"SSHService", ps.createSSHService},
		{"SSHAgentService", ps.createSSHAgentService},
	}
	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			for _, tc := range []struct {
				listen  string
				wantErr bool
			}{
				{"tcp://localhost:0", false},
				{"tcp://bad:0", true},
			} {
				cfg := v2.Service{
					Name:      tt.name,
					Connector: "mock",
					ListenOn:  v2.NetworkAddress(tc.listen),
				}
				ps.availPlugins = &fakeAvailablePlugins{}
				svc, err := tt.createFn(cfg)
				if tc.wantErr {
					assert.Error(t, err)
					assert.Nil(t, svc)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, svc)
				}
			}
		})
	}
}

// --- Helpers & Mocks ---

type mockService struct {
	stopped bool
	stopErr error
}

func (m *mockService) Start() error { return nil }
func (m *mockService) Stop() error  { m.stopped = true; return m.stopErr }

type mockResolver struct {
	called bool
}

func (m *mockResolver) Resolve(_ []*v2.Credential) (map[string][]byte, error) {
	m.called = true
	return map[string][]byte{"foo": []byte("bar")}, nil
}
func (m *mockResolver) Provider(string) (v1.Provider, error) { return nil, nil }

type testLogger struct{}

func (l *testLogger) CopyWith(name string, debug bool) logapi.Logger { return l }
func (l *testLogger) DebugEnabled() bool                             { return false }
func (l *testLogger) Prefix() string                                 { return "" }
func (l *testLogger) Debug(args ...interface{})                      {}
func (l *testLogger) Debugf(format string, args ...interface{})      {}
func (l *testLogger) Debugln(args ...interface{})                    {}
func (l *testLogger) Info(args ...interface{})                       {}
func (l *testLogger) Infof(format string, args ...interface{})       {}
func (l *testLogger) Infoln(args ...interface{})                     {}
func (l *testLogger) Warn(args ...interface{})                       {}
func (l *testLogger) Warnf(format string, args ...interface{})       {}
func (l *testLogger) Warnln(args ...interface{})                     {}
func (l *testLogger) Error(args ...interface{})                      {}
func (l *testLogger) Errorf(format string, args ...interface{})      {}
func (l *testLogger) Errorln(args ...interface{})                    {}
func (l *testLogger) Panic(args ...interface{})                      {}
func (l *testLogger) Panicf(format string, args ...interface{})      {}
func (l *testLogger) Panicln(args ...interface{})                    {}

type fakeTCPPlugin struct{}

func (f *fakeTCPPlugin) NewConnector(_ connector.Resources) tcpplugin.Connector {
	return &fakeTCPConnector{}
}

type fakeTCPConnector struct{}

func (f *fakeTCPConnector) Connect(_ net.Conn, _ connector.CredentialValuesByID) (net.Conn, error) {
	return nil, nil
}

type fakeAvailablePlugins struct{}

func (f *fakeAvailablePlugins) HTTPPlugins() map[string]httpplugin.Plugin { return nil }
func (f *fakeAvailablePlugins) TCPPlugins() map[string]tcpplugin.Plugin {
	return map[string]tcpplugin.Plugin{"mock": &fakeTCPPlugin{}}
}

type fakeAvailablePluginsWithHTTPMock struct{}

func (f *fakeAvailablePluginsWithHTTPMock) HTTPPlugins() map[string]httpplugin.Plugin {
	return map[string]httpplugin.Plugin{"mock": &httpPluginParseFail{}}
}
func (f *fakeAvailablePluginsWithHTTPMock) TCPPlugins() map[string]tcpplugin.Plugin { return nil }

type httpPluginParseFail struct{}

func (h *httpPluginParseFail) NewConnector(_ connector.Resources) httpplugin.Connector {
	return &httpConnectorParseFail{}
}

type httpConnectorParseFail struct{}

func (h *httpConnectorParseFail) Connect(_ *http.Request, _ connector.CredentialValuesByID) error {
	return errors.New("parse fail")
}
