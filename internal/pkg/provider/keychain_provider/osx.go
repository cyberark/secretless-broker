// +build darwin

package keychain_provider

// See https://github.com/keybase/go-keychain/blob/master/keychain.go

// See https://developer.apple.com/library/ios/documentation/Security/Reference/keychainservices/index.html for the APIs used below.

// Also see https://developer.apple.com/library/ios/documentation/Security/Conceptual/keychainServConcepts/01introduction/introduction.html .

/*
#cgo LDFLAGS: -framework CoreFoundation -framework Security
#include <stdlib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func release(ref C.CFTypeRef) {
	C.CFRelease(ref)
}

func GetGenericPassword(service, account string) ([]byte, error) {
	service_c := C.CString(service)
	account_c := C.CString(account)

	for _, ptr := range []*C.char{service_c, account_c} {
		defer C.free(unsafe.Pointer(ptr))
	}

	service_cf := C.CFStringCreateWithCString(nil, service_c, C.kCFStringEncodingUTF8)
	account_cf := C.CFStringCreateWithCString(nil, account_c, C.kCFStringEncodingUTF8)

	for _, ptr := range []C.CFStringRef{service_cf, account_cf} {
		defer release(C.CFTypeRef(ptr))
	}

	keys := make([]C.CFTypeRef, 4)
	values := make([]C.CFTypeRef, 4)

	keys[0] = C.CFTypeRef(C.kSecAttrService)
	keys[1] = C.CFTypeRef(C.kSecAttrAccount)
	keys[2] = C.CFTypeRef(C.kSecClass)
	keys[3] = C.CFTypeRef(C.kSecReturnData)

	values[0] = C.CFTypeRef(service_cf)
	values[1] = C.CFTypeRef(account_cf)
	values[2] = C.CFTypeRef(C.kSecClassGenericPassword)
	values[3] = C.CFTypeRef(C.kCFBooleanTrue)

	keyCallbacks := (*C.CFDictionaryKeyCallBacks)(&C.kCFTypeDictionaryKeyCallBacks)
	valCallbacks := (*C.CFDictionaryValueCallBacks)(&C.kCFTypeDictionaryValueCallBacks)

	query_cf := C.CFDictionaryCreate(nil, (*unsafe.Pointer)(&keys[0]), (*unsafe.Pointer)(&values[0]), C.CFIndex(len(keys)), keyCallbacks, valCallbacks)

	defer release(C.CFTypeRef(query_cf))

	var resultsRef C.CFTypeRef
	errCode := C.SecItemCopyMatching(query_cf, &resultsRef)
	if errCode != 0 {
		errorMessage_cf := C.SecCopyErrorMessageString(errCode, nil)
		defer release(C.CFTypeRef(errorMessage_cf))
		// Whether or not this function returns a valid pointer or NULL depends on many factors, ...
		errorMessage_c := C.CFStringGetCStringPtr(errorMessage_cf, C.kCFStringEncodingUTF8)
		var message string
		if errorMessage_c != nil {
			message = C.GoString(errorMessage_c)
		} else {
			C.CFShow(C.CFTypeRef(errorMessage_cf))
			message = fmt.Sprintf("An unknown error occurred : %d", int(errCode))
		}
		return nil, fmt.Errorf(message)
	}

	defer release(C.CFTypeRef(resultsRef))

	cfData := C.CFDataRef(resultsRef)
	bytes := C.GoBytes(unsafe.Pointer(C.CFDataGetBytePtr(cfData)), C.int(C.CFDataGetLength(cfData)))

	return bytes, nil
}
