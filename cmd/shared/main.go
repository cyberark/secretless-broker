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
	"github.com/cyberark/secretless-broker/internal/pkg/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	"crypto/sha1"
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

func NewStoredSecret(ref C.struct_StoredSecret) config.StoredSecret {
	return config.StoredSecret{
		Name:     C.GoString(ref.Name),
		ID:       C.GoString(ref.ID),
		Provider: C.GoString(ref.Provider),
	}
}

func GetSecrets(secrets []config.StoredSecret) (map[string][]byte, error)  {
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
	secrets, err := GetSecrets([]config.StoredSecret{ref})
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

	sha1 := sha1.New()
	sha1.Write(passwordBytes)
	passwordSHA1 := sha1.Sum(nil)

	sha1.Reset()
	sha1.Write(passwordSHA1)
	hash := sha1.Sum(nil)

	sha1.Reset()

	sha1.Write(saltBytes)
	sha1.Write(hash)
	randomSHA1 := sha1.Sum(nil)

	// nativePassword = passwordSHA1 ^ randomSHA1
	nativePassword := make([]byte, len(randomSHA1))
	defer ZeroizeByteSlice(nativePassword)
	for i := range randomSHA1 {
		nativePassword[i] = passwordSHA1[i] ^ randomSHA1[i]
	}

	return C.CString(ByteBoundString(nativePassword))
}

func main() {}
