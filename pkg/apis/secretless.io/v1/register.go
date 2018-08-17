package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	secretlessio "github.com/cyberark/secretless-broker/pkg/apis/secretless.io"
)

// SchemeGroupVersion indicates the combo of group name and version that we are
// defining these schemes for
var SchemeGroupVersion = schema.GroupVersion{
	// Group is the CRD TLD identifier
	Group: secretlessio.GroupName,

	// Version indicates the SemVer of this CRD object definition
	Version: "v1",
}

// Resource returns a group resource of our specificed group/version combo with a
// specific resource ID
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder is a parent object that we copy fields from
	SchemeBuilder runtime.SchemeBuilder

	localSchemeBuilder = &SchemeBuilder

	// AddToScheme indicates the function that will be used to add types
	AddToScheme = localSchemeBuilder.AddToScheme
)

func init() {
	localSchemeBuilder.Register(addKnownTypes)
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Configuration{},
		&ConfigurationList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)

	return nil
}
