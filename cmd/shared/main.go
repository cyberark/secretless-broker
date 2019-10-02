package main

/*
struct StoredSecret{
    char *Name;
    char *ID;
    char *Provider;
};
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	secretless "github.com/cyberark/secretless-broker/internal"
	"github.com/cyberark/secretless-broker/internal/plugin"
	pluginv1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/proxyservice/tcp/mysql/protocol"
	configv2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// ZeroizeByteSlice sets every byte to zero.
func ZeroizeByteSlice(bs []byte) {
	for byteIndex := range bs {
		bs[byteIndex] = 0
	}
}

// ByteBoundString returns a string backed by the given []byte.
func ByteBoundString(b []byte) string {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bytesHeader := &reflect.StringHeader{
		Data: header.Data,
		Len: header.Len,
	}
	return *(*string)(unsafe.Pointer(bytesHeader))
}

// NewCredential create a Credential from the given C struct.
func NewCredential(ref C.struct_StoredSecret) *configv2.Credential {
	return &configv2.Credential{
		Name:     C.GoString(ref.Name),
		Get:       C.GoString(ref.ID),
		From: C.GoString(ref.Provider),
	}
}

// GetSecrets returns secret values.  Specifically, a map whose keys are the
// secret names requested, and whose values are the values of those secrets.
func GetSecrets(secrets []*configv2.Credential) (map[string][]byte, error)  {
	// Load all internal Providers
	providerFactories := make(map[string]func(pluginv1.ProviderOptions) (pluginv1.Provider, error))
	for providerID, providerFactory := range secretless.InternalProviders {
		providerFactories[providerID] = providerFactory
	}

	resolver := plugin.NewResolver(providerFactories, nil, nil)

	return resolver.Resolve(secrets)
}

// GetSecret returns a C *char with the given secret's value
// export GetSecret
func GetSecret(cRef C.struct_StoredSecret) (*C.char) {
	return C.CString(GetSecretByteString(cRef))
}

// GetSecretByteString return the secret value for the given StoredSecret ref.
func GetSecretByteString(cRef C.struct_StoredSecret) (string) {
	ref := NewCredential(cRef)
	secrets, err := GetSecrets([]*configv2.Credential{ref})
	if err != nil {
		fmt.Println("Error fetching secret")
		return ByteBoundString(nil)
	}
	return ByteBoundString(secrets[ref.Name])
}

// NativePassword returns the given StoredSecret value as C *char.
// export NativePassword
func NativePassword(cRef C.struct_StoredSecret, salt *C.char) (*C.char) {
	passwordBytes := []byte(GetSecretByteString(cRef))
	defer ZeroizeByteSlice(passwordBytes)
	saltBytes := C.GoBytes(unsafe.Pointer(salt), C.int(8))
	defer ZeroizeByteSlice(saltBytes)

	// nativePassword = passwordSHA1 ^ randomSHA1
	nativePassword, _ := protocol.NativePassword(passwordBytes, saltBytes)
	defer ZeroizeByteSlice(nativePassword)

	return C.CString(ByteBoundString(nativePassword))
}

func main() {}
