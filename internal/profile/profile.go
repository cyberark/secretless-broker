package profile

import (
	"fmt"

	"github.com/pkg/profile"
)

// ProfilingService is the interface to this class' functionality.
type ProfilingService interface {
	Start()
	Stop()
}

// StartProfilingFunc is a function that will be invoked when profile start
// is requested. This method signature is just a type wrapper around the
// `Start()` signature from `github.com/pkg/profile`.
type StartProfilingFunc func(...func(*profile.Profile)) interface{ Stop() }

// Start is used to begin the profiling
func (pp *perfProfile) Start() {
	pp.stopper = pp.startProfiling(pp.profileOptions...)
}

// Start is used to end the profiling
func (pp *perfProfile) Stop() {
	pp.stopper.Stop()
}

type perfProfile struct {
	profileOptions []func(*profile.Profile)
	startProfiling StartProfilingFunc
	stopper        interface{ Stop() }
}

// validTypes are the valid types of profiling you can perform.
var validTypes = []string{"cpu", "memory"}

func isValidType(profileType string) bool {
	for _, curType := range validTypes {
		if curType == profileType {
			return true
		}
	}
	return false
}

// While it might make sense to make this validation public and allow it to
// occur at a higher level, we decided to keep it here because:
// 1. it was so simple and
// 2. this code is being called directly from the entrypoint anyway, and so
//    is already close to the border.
func validateType(profileType string) error {
	if !isValidType(profileType) {
		return fmt.Errorf(
			"Invalid profile type: '%s'.  Valid types are: %v",
			profileType,
			validTypes,
		)
	}
	return nil
}

func profileOptionsForType(profileType string) []func(*profile.Profile) {
	switch profileType {
	case "cpu":
		return []func(*profile.Profile){profile.CPUProfile, profile.NoShutdownHook}
	case "memory":
		return []func(*profile.Profile){profile.MemProfile, profile.NoShutdownHook}
	}

	return []func(*profile.Profile){}
}

// New returns a new performance profile of the specified type.
func New(profileType string) (ProfilingService, error) {
	return NewWithOptions(profileType, profile.Start)
}

// NewWithOptions returns a new performance profile of the specified type
// with the ability to pass in the implementing profile interface..
func NewWithOptions(
	profileType string,
	startProfilingFunc StartProfilingFunc,
) (ProfilingService, error) {

	err := validateType(profileType)
	if err != nil {
		return nil, err
	}

	return &perfProfile{
		profileOptions: profileOptionsForType(profileType),
		startProfiling: startProfilingFunc,
	}, nil
}
