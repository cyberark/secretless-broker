package main

/*
struct CredentialSpec{
    char *Name;
    char *Get;
    char *From;
};
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/cyberark/secretless-broker/internal/plugin"
	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mysql/protocol"
	pluginv1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
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
		Len:  header.Len,
	}
	return *(*string)(unsafe.Pointer(bytesHeader))
}

// NewCredential creates a Credential from the given C struct.
func NewCredential(ref C.struct_CredentialSpec) *configv2.Credential {
	return &configv2.Credential{
		Name: C.GoString(ref.Name),
		Get:  C.GoString(ref.ID),
		From: C.GoString(ref.Provider),
	}
}

// GetCredentialValues returns credential values.  Specifically, a map whose keys are the
// credential IDs requested, and whose values are the values of those credentials.
func GetCredentialValues(credentialSpecs []*configv2.Credential) (map[string][]byte, error) {
	// Load all internal Providers
	providerFactories := make(map[string]func(pluginv1.ProviderOptions) (pluginv1.Provider, error))
	for providerID, providerFactory := range providers.ProviderFactories {
		providerFactories[providerID] = providerFactory
	}

	resolver := plugin.NewResolver(providerFactories, nil, nil)

	return resolver.Resolve(credentialSpecs)
}

// GetCredentialValue returns a C *char with the given credential's value
// export GetCredentialValue
func GetCredentialValue(cRef C.struct_CredentialSpec) *C.char {
	return C.CString(GetCredentialValueByteString(cRef))
}

// GetCredentialValueByteString return the credential value for the given CredentialSpec ref.
func GetCredentialValueByteString(cRef C.struct_CredentialSpec) string {
	ref := NewCredential(cRef)
	credentials, err := GetCredentialValues([]*configv2.Credential{ref})
	if err != nil {
		fmt.Println("Error fetching credential")
		return ByteBoundString(nil)
	}
	return ByteBoundString(credentials[ref.Name])
}

// NativePassword returns the given CredentialSpec value as C *char.
// export NativePassword
func NativePassword(cRef C.struct_CredentialSpec, salt *C.char) *C.char {
	passwordBytes := []byte(GetCredentialValueByteString(cRef))
	defer ZeroizeByteSlice(passwordBytes)
	saltBytes := C.GoBytes(unsafe.Pointer(salt), C.int(8))
	defer ZeroizeByteSlice(saltBytes)

	// nativePassword = passwordSHA1 ^ randomSHA1
	nativePassword, _ := protocol.NativePassword(passwordBytes, saltBytes)
	defer ZeroizeByteSlice(nativePassword)

	return C.CString(ByteBoundString(nativePassword))
}

func main() {}
