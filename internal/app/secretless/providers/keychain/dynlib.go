package keychain

// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include <dlfcn.h>
import "C"

import (
	"errors"
	"fmt"
	"log"
	"unsafe"
)

type DynamicLibrary struct {
	Name string
	Ref  unsafe.Pointer
}

type Symbol struct {
	Name string
	Ref  unsafe.Pointer
}

func OpenDynamicLibrary(libraryName string) (*DynamicLibrary, error) {
	log.Printf("Opening dynamic library '%s'...", libraryName)

	libraryNameCStr := C.CString(libraryName)
	defer C.free(unsafe.Pointer(libraryNameCStr))

	// Lazy load the library
	libraryRef := C.dlopen(libraryNameCStr, C.RTLD_LAZY)

	if libraryRef == nil {
		return nil, fmt.Errorf("ERROR: Unable to find dynamic library '%s'!", libraryName)
	}

	log.Printf("Opened dynamic library '%s'", libraryName)

	return &DynamicLibrary{
		Name: libraryName,
		Ref:  libraryRef,
	}, nil
}

func (dynlib *DynamicLibrary) Close() error {
	log.Printf("Closing dynamic library '%s'...", dynlib.Name)

	// Clear error state
	C.dlerror()

	C.dlclose(dynlib.Ref)
	if dlErr := C.dlerror(); dlErr != nil {
		return fmt.Errorf("ERROR: Unable to close %s: %s!", dynlib.Name, errors.New(C.GoString(dlErr)))
	}

	log.Printf("Closed dynamic library '%s'.", dynlib.Name)

	return nil
}

func (dynlib *DynamicLibrary) GetSymbol(symbolName string) (*Symbol, error) {
	log.Printf("Getting pointer to '%s:%s'...", dynlib.Name, symbolName)

	if symbolName == "" {
		return nil, fmt.Errorf("ERROR: Symbol name provided is blank!")
	}

	symbolNameCStr := C.CString(symbolName)
	defer C.free(unsafe.Pointer(symbolNameCStr))

	// Clear error state
	C.dlerror()

	symbolRef := C.dlsym(dynlib.Ref, symbolNameCStr)
	if symErr := C.dlerror(); symErr != nil {
		return nil, fmt.Errorf("ERROR: Unable to resolve '%s:%s'!", symbolName, errors.New(C.GoString(symErr)))
	}

	log.Printf("Found pointer to '%s': %v", symbolName, symbolRef)

	return &Symbol{
		Name: symbolName,
		Ref:  symbolRef,
	}, nil
}
