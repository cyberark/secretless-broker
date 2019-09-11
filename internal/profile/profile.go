package profile

import (
	"fmt"
	"strings"

	"github.com/pkg/profile"
)

func (pp *perfProfile) Start() {
	switch pp.profileType {
	case "cpu":
		pp.stopper = profile.Start(profile.NoShutdownHook)
	case "memory":
		pp.stopper = profile.Start(profile.MemProfile, profile.NoShutdownHook)
	default:
		// will be impossible when New is used as ctor
		panic("Attempt to start profiling with invalid profileType")
	}
}

func (pp *perfProfile) Stop() {
	pp.stopper.Stop()
}

type perfProfile struct {
	profileType string
	stopper interface { Stop() }
}

var ValidTypes = []string{"cpu", "memory"}

func isValidType(profileType string) bool {
	for _, curType := range ValidTypes {
		if curType == profileType {
			return true
		}
	}
	return false
}

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

func New(profileType string) *perfProfile {
	// Clients are expected to have validated the type
	if !isValidType(profileType) {
		panic("profile type must be 'cpu' or 'memory'")
	}
	return &perfProfile{ profileType: profileType }
}
