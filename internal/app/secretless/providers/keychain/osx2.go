// +build darwin

package keychain

// #include <stdlib.h>
// #include <stdint.h>
// #include <stdio.h>
// #include <strings.h>
//
// const void*
// CFStringCreateWithCStringSymbol_wrapper(void *func_ptr, const char *c_str)
// {
//   typedef uint32_t CFStringEncoding;
//   const uint32_t kCFStringEncodingUTF8 = 0x08000100;
//   const void* (*cfStringCreateWithCStringSymbol)(int, const char *, CFStringEncoding) = (const void* (*)(int, const char *, CFStringEncoding))func_ptr;
//   const void *result = cfStringCreateWithCStringSymbol(0, c_str, kCFStringEncodingUTF8);
//   printf("%p\n", result);
//   return result;
// }
//
// const char*
// SecKeychainFindGenericPassword_wrapper(void *func_ptr, const char *account, const char *service, int *length, void **data)
// {
//   typedef const void* func_type(uintptr_t, uint32_t, const char *, uint32_t, const char *, int *, uintptr_t, uintptr_t);
//   uint32_t password_length = 0;
//   uintptr_t password_data = ((void *)0);
//   const void* (*func_cast)(uintptr_t,
//                            uint32_t,
//                            const char *,
//                            uint32_t,
//                            const char *,
//                            uint32_t *,
//                            uintptr_t,
//                            uintptr_t) = (func_type *) func_ptr;
//
//   printf("Stuff: %s\n", service);
//   printf("Stuff: %s\n", account);
//   const int ret = func_cast(0,
//                             strlen(service),
//                             service,
//                             strlen(account),
//                             account,
//                             &password_length,
//                             &password_data,
//                             0);
//
//   *length = password_length;
//   *data = password_data;
//
//   printf("Sec result: %d\n", ret);
//   printf("PW length: %d\n", password_length);
//   printf("PW: %.*s\n", password_length, (const char *) password_data);
//   return (const char *)password_data;
// }
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

type SecurityItemQuery struct {
	Account      string
	Service      string
	DynlibSymbol *Symbol
}

func (secQuery *SecurityItemQuery) Execute() ([]byte, error) {
	const NULL = 0

	libName := "CoreFoundation"
	libPath := fmt.Sprintf("/System/Library/Frameworks/%s.framework/Versions/A/%s", libName, libName)

	log.Println("Executing SecItemCopyMatching query...")

	dynamicLibrary, err := OpenDynamicLibrary(libPath)
	if err != nil {
		return nil, err
	}
	defer dynamicLibrary.Close()

	// Get our helper symbols loaded
	cfStringCreateWithCStringSymbol, err := dynamicLibrary.GetSymbol("CFStringCreateWithCString")
	if err != nil {
		return nil, err
	}

	// Turn our Golang strings into C strings
	accountNameCStr := C.CString(secQuery.Account)
	serviceNameCStr := C.CString(secQuery.Service)
	defer C.free(unsafe.Pointer(accountNameCStr))
	defer C.free(unsafe.Pointer(serviceNameCStr))

	// Turn those C strings into OSX-specific immutable strings
	accountNameCFStr := C.CFStringCreateWithCStringSymbol_wrapper(cfStringCreateWithCStringSymbol.Ref,
		accountNameCStr)
	serviceNameCFStr := C.CFStringCreateWithCStringSymbol_wrapper(cfStringCreateWithCStringSymbol.Ref,
		serviceNameCStr)
	defer C.free(unsafe.Pointer(accountNameCFStr))
	defer C.free(unsafe.Pointer(serviceNameCFStr))

	log.Printf("accountNameCFStr: %v", accountNameCFStr)
	log.Printf("serviceNameCFStr: %v", serviceNameCFStr)

	log.Println("Retrieved credential(s) from SecItemCopyMatching query")

	var credentialLength C.int
	var credentialData unsafe.Pointer

	foo := C.SecKeychainFindGenericPassword_wrapper(secQuery.DynlibSymbol.Ref,
		accountNameCStr,
		serviceNameCStr,
		&credentialLength,
		&credentialData)

	defer C.free(unsafe.Pointer(credentialData))

	log.Printf("A: %v", credentialLength)
	log.Printf("B: %v", credentialData)
	log.Printf("Result: %v", foo)

	bytes := C.GoBytes(unsafe.Pointer(credentialData), credentialLength)

	return bytes, nil
}

// GetGenericPassword returns password data for service and account
func GetGenericPassword(service string, account string) ([]byte, error) {
	log.Println("Processing credentials...")
	libName := "Security"
	libPath := fmt.Sprintf("/System/Library/Frameworks/%s.framework/Versions/A/%s", libName, libName)

	dynamicLibrary, err := OpenDynamicLibrary(libPath)
	if err != nil {
		return nil, err
	}
	defer dynamicLibrary.Close()

	//symbol, err := dynamicLibrary.GetSymbol("SecItemCopyMatching")
	symbol, err := dynamicLibrary.GetSymbol("SecKeychainFindGenericPassword")
	if err != nil {
		return nil, err
	}

	secItemQuery := &SecurityItemQuery{
		DynlibSymbol: symbol,
		Account:      account,
		Service:      service,
	}

	return secItemQuery.Execute()
}
