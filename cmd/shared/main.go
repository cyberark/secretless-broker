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
	"github.com/cyberark/secretless-broker/internal/app/secretless"
	"github.com/cyberark/secretless-broker/internal/app/secretless/handlers/mysql/protocol"
	"github.com/cyberark/secretless-broker/internal/pkg/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"reflect"
	"unsafe"
)

func ZeroizeByteSlice(bs []byte) {
	for byteIndex := range bs {
		bs[byteIndex] = 0
	}
}

func ByteBoundString(b []byte) string {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bytesHeader := &reflect.StringHeader{
		Data: header.Data,
		Len: header.Len,
	}
	return *(*string)(unsafe.Pointer(bytesHeader))
}

func NewStoredSecret(ref C.struct_StoredSecret) v1.StoredSecret {
	return v1.StoredSecret{
		Name:     C.GoString(ref.Name),
		ID:       C.GoString(ref.ID),
		Provider: C.GoString(ref.Provider),
	}
}

func GetSecrets(secrets []v1.StoredSecret) (map[string][]byte, error)  {
	// Load all internal Providers
	providerFactories := make(map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error))
	for providerID, providerFactory := range secretless.InternalProviders {
		providerFactories[providerID] = providerFactory
	}

	resolver := plugin.NewResolver(providerFactories, nil, nil)

	return resolver.Resolve(secrets)
}

//export GetSecret
func GetSecret(cRef C.struct_StoredSecret) (*C.char) {
	return C.CString(GetSecretByteString(cRef))
}

func GetSecretByteString(cRef C.struct_StoredSecret) (string) {
	ref := NewStoredSecret(cRef)
	secrets, err := GetSecrets([]v1.StoredSecret{ref})
	if err != nil {
		fmt.Println("Error fetching secret")
		return ByteBoundString(nil)
	}
	return ByteBoundString(secrets[ref.Name])
}

//export NativePassword
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
