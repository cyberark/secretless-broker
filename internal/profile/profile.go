package profile

import (
	"fmt"
	"strings"

	"github.com/cyberark/secretless-broker/internal"
	"github.com/pkg/profile"
)

func (pp *perfProfile) Start() error {
	switch pp.profileType {
	case "cpu":
		pp.stopper = profile.Start(profile.NoShutdownHook)
	case "memory":
		pp.stopper = profile.Start(profile.MemProfile, profile.NoShutdownHook)
	default:
		// will be impossible when New is used as ctor
		return fmt.Errorf("attempt to start profiling with invalid profileType")
	}
	return nil
}

func (pp *perfProfile) Stop() error {
	pp.stopper.Stop()
	return nil
}

type perfProfile struct {
	profileType string
	stopper interface { Stop() }
}

// ValidTypes are the valid types of profiling you can perform.
var ValidTypes = []string{"cpu", "memory"}

func isValidType(profileType string) bool {
	for _, curType := range ValidTypes {
		if curType == profileType {
			return true
		}
	}
	return false
}

// ValidateType returns an error unless its argument is a valid profile type.
func ValidateType(profileType string) error {
	if !isValidType(profileType) {
		return fmt.Errorf(
			"Invalid profile type: '%s'.  Valid types are: %s",
			profileType,
			strings.Join(ValidTypes, ", "),
		)
	}
	return nil
}

// New returns a new performance profile of the specified type.
func New(profileType string) internal.Service {
	// Clients are expected to have validated the type
	if !isValidType(profileType) {
		panic("profile type must be 'cpu' or 'memory'")
	}
	return &perfProfile{ profileType: profileType }
}
