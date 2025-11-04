package conjur

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/stretchr/testify/assert"
)

type mockConjurClient struct{}

func (m *mockConjurClient) RetrieveSecret(id string) ([]byte, error) {
	return []byte("dummy"), nil
}

func TestProviderFactory_TokenFile(t *testing.T) {
	os.Setenv("CONJUR_AUTHN_TOKEN_FILE", "/tmp/token")
	os.Setenv("CONJUR_ACCOUNT", "dev")
	os.Setenv("CONJUR_APPLIANCE_URL", "http://conjur")

	f, _ := os.Create("/tmp/token")
	f.Close()

	provider, err := ProviderFactory(plugin_v1.ProviderOptions{Name: "test"})
	assert.NoError(t, err)
	assert.Equal(t, "test", provider.GetName())
	os.Remove("/tmp/token")
}

func TestProviderFactory_Error(t *testing.T) {
	os.Unsetenv("CONJUR_AUTHN_LOGIN")
	os.Unsetenv("CONJUR_AUTHN_API_KEY")
	os.Unsetenv("CONJUR_AUTHN_TOKEN_FILE")
	os.Unsetenv("CONJUR_AUTHN_URL")

	_, err := ProviderFactory(plugin_v1.ProviderOptions{Name: "test"})
	assert.Error(t, err)
}

func TestProvider_GetName(t *testing.T) {
	p := &Provider{Name: "abc"}
	assert.Equal(t, "abc", p.GetName())
}

func TestProvider_GetValues(t *testing.T) {
	p := &Provider{
		Conjur: &conjurapi.Client{},
	}
	_, err := p.GetValues("id1", "id2")
	assert.NoError(t, err)
}

func TestProvider_GetValue_AccessToken_NoClient(t *testing.T) {
	p := &Provider{
		Username: "user",
		APIKey:   "key",
		Conjur:   &conjurapi.Client{},
	}
	assert.Panics(t, func() {
		_, _ = p.GetValue("accessToken")
	})
}

func TestProvider_GetValue_Error(t *testing.T) {
	p := &Provider{}
	_, err := p.GetValue("accessToken")
	assert.Error(t, err)
}

func TestProvider_GetValue_Variable(t *testing.T) {
	p := &Provider{
		Config: conjurapi.Config{Account: "acc"},
		Conjur: &conjurapi.Client{},
	}
	_, err := p.GetValue("varid")
	assert.Error(t, err)
}

func TestProvider_fetchAccessTokenLoop_NilAuthenticator(t *testing.T) {
	p := &Provider{}
	err := p.fetchAccessTokenLoop()
	assert.Error(t, err)
}

func Test_urlSupported(t *testing.T) {
	assert.True(t, urlSupported("authn-k8s/abc"))
	assert.True(t, urlSupported("authn-jwt/abc"))
	assert.False(t, urlSupported(""))
	assert.False(t, urlSupported("other"))
}

type mockTokenTimeout struct {
	timeout time.Duration
}

func (m *mockTokenTimeout) GetTokenTimeout() time.Duration {
	return m.timeout
}

func TestProviderFactory_AuthenticatorConfigError(t *testing.T) {
	// Simulate urlSupported returns true, but config fails
	os.Setenv("CONJUR_AUTHN_URL", "authn-k8s/abc")
	os.Unsetenv("CONJUR_AUTHN_LOGIN")
	os.Unsetenv("CONJUR_AUTHN_API_KEY")
	os.Unsetenv("CONJUR_AUTHN_TOKEN_FILE")
	// Unset required env for config to fail
	os.Unsetenv("CONJUR_ACCOUNT")
	_, err := ProviderFactory(plugin_v1.ProviderOptions{Name: "test"})
	assert.Error(t, err)
	os.Unsetenv("CONJUR_AUTHN_URL")
}

func TestProviderFactory_UnsupportedURL(t *testing.T) {
	os.Setenv("CONJUR_AUTHN_URL", "other")
	_, err := ProviderFactory(plugin_v1.ProviderOptions{Name: "test"})
	assert.Error(t, err)
	os.Unsetenv("CONJUR_AUTHN_URL")
}

type mockAuthenticatorSuccess struct {
	mu     sync.Mutex
	called bool
}

type mockAuthenticatorOnce struct {
	mu    sync.Mutex
	calls int
}

func (m *mockAuthenticatorSuccess) Authenticate() error {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()
	return nil
}

func (m *mockAuthenticatorSuccess) AuthenticateWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.Authenticate()
	}
}

func (m *mockAuthenticatorSuccess) GetAccessToken() access_token.AccessToken {
	var t access_token.AccessToken
	return t
}

func (m *mockAuthenticatorOnce) GetAccessToken() access_token.AccessToken {
	var t access_token.AccessToken
	return t
}

func (m *mockAuthenticatorOnce) Authenticate() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	if m.calls == 1 {
		return errors.New("temporary error")
	}
	return nil
}

func (m *mockAuthenticatorOnce) AuthenticateWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.Authenticate()
	}
}

func Test_fetchAccessToken_Succeeds(t *testing.T) {
	m := &mockAuthenticatorSuccess{}
	p := &Provider{
		AuthenticationMutex: &sync.Mutex{},
		Authenticator:       m,
	}
	err := p.fetchAccessToken()
	assert.NoError(t, err)
	assert.True(t, m.called, "Authenticate should have been called")
}
