// +build linux

package keychain_provider

import (
	"fmt"
)

func GetGenericPassword(service, account string) ([]byte, error) {
	return nil, fmt.Errorf("No keychain provider for Linux (yet)")
}
