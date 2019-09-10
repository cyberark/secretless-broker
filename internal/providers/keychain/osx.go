// +build darwin

package keychain

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

// NULL is just an int representation of C NULL
const NULL = 0

func release(ref C.CFTypeRef) {
	C.CFRelease(ref)
}

// GetGenericPassword returns password data for service and account
func GetGenericPassword(service string, account string) ([]byte, error) {
	serviceC := C.CString(service)
	accountC := C.CString(account)

	for _, ptr := range []*C.char{serviceC, accountC} {
		defer C.free(unsafe.Pointer(ptr))
	}

	serviceCf := C.CFStringCreateWithCString(NULL, serviceC, C.kCFStringEncodingUTF8)
	accountCf := C.CFStringCreateWithCString(NULL, accountC, C.kCFStringEncodingUTF8)

	for _, ptr := range []C.CFStringRef{serviceCf, accountCf} {
		defer release(C.CFTypeRef(ptr))
	}

	keys := make([]C.CFTypeRef, 4)
	values := make([]C.CFTypeRef, 4)

	keys[0] = C.CFTypeRef(C.kSecAttrService)
	keys[1] = C.CFTypeRef(C.kSecAttrAccount)
	keys[2] = C.CFTypeRef(C.kSecClass)
	keys[3] = C.CFTypeRef(C.kSecReturnData)

	values[0] = C.CFTypeRef(serviceCf)
	values[1] = C.CFTypeRef(accountCf)
	values[2] = C.CFTypeRef(C.kSecClassGenericPassword)
	values[3] = C.CFTypeRef(unsafe.Pointer(C.kCFBooleanTrue))

	keyCallbacks := (*C.CFDictionaryKeyCallBacks)(&C.kCFTypeDictionaryKeyCallBacks)
	valCallbacks := (*C.CFDictionaryValueCallBacks)(&C.kCFTypeDictionaryValueCallBacks)

	queryCf := C.CFDictionaryCreate(NULL, (*unsafe.Pointer)(unsafe.Pointer(&keys[0])), (*unsafe.Pointer)(unsafe.Pointer(&values[0])), C.CFIndex(len(keys)), keyCallbacks, valCallbacks)

	defer release(C.CFTypeRef(unsafe.Pointer(queryCf)))

	var resultsRef C.CFTypeRef
	errCode := C.SecItemCopyMatching(queryCf, &resultsRef)
	if errCode != 0 {
		errorMessageCf := C.SecCopyErrorMessageString(errCode, nil)
		defer release(C.CFTypeRef(errorMessageCf))
		// Whether or not this function returns a valid pointer or NULL depends on many factors, ...
		errorMessageC := C.CFStringGetCStringPtr(errorMessageCf, C.kCFStringEncodingUTF8)
		var message string
		if errorMessageC != nil {
			message = C.GoString(errorMessageC)
		} else {
			C.CFShow(C.CFTypeRef(errorMessageCf))
			message = fmt.Sprintf("An unknown error occurred : %d", int(errCode))
		}
		return nil, fmt.Errorf(message)
	}

	defer release(C.CFTypeRef(resultsRef))

	cfData := C.CFDataRef(resultsRef)
	bytes := C.GoBytes(unsafe.Pointer(C.CFDataGetBytePtr(cfData)), C.int(C.CFDataGetLength(cfData)))

	return bytes, nil
}
