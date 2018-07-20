// +build linux

package keychain

import (
	"fmt"
)

// GetGenericPassword returns password data for service and account
func GetGenericPassword(service, account string) ([]byte, error) {
	return nil, fmt.Errorf("No keychain provider for Linux (yet)")
}
