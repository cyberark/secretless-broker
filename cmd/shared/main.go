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
)

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
	ref := NewStoredSecret(cRef)
	secrets, err := GetSecrets([]config.StoredSecret{ref})
	if err != nil {
		fmt.Println("Error fetching secret")
		return C.CString("")
	}
	return C.CString(string(secrets[ref.Name]))
}

//export NativePassword
func NativePassword(cRef C.struct_StoredSecret, salt *C.char) (*C.char) {
	password := GetSecret(cRef)

	sha1 := sha1.New()
	sha1.Write([]byte(C.GoString(password)))
	passwordSHA1 := sha1.Sum(nil)

	sha1.Reset()
	sha1.Write(passwordSHA1)
	hash := sha1.Sum(nil)

	sha1.Reset()
	sha1.Write([]byte(C.GoString(salt)))
	sha1.Write(hash)
	randomSHA1 := sha1.Sum(nil)

	// nativePassword = passwordSHA1 ^ randomSHA1
	nativePassword := make([]byte, len(randomSHA1))
	for i := range randomSHA1 {
		nativePassword[i] = passwordSHA1[i] ^ randomSHA1[i]
	}

	return C.CString(string(nativePassword))
}

func main() {}
