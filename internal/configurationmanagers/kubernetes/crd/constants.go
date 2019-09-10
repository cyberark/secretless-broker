package crd

import (
	"os"
	"time"
)

var (
	// CRDGroupName is the main interface TLD that we tie our CRD under
	CRDGroupName = "secretless" + os.Getenv("SECRETLESS_CRD_SUFFIX") + ".io"

	// CRDFQDNName is the fully-qualified resource ID
	CRDFQDNName = CRDName + "." + CRDGroupName
)

const (

	// CRDLongName is a string indicating what prefix we will use on the CLI
	CRDLongName = "configuration"

	// CRDName is the internal prefix for our resource that will be prefixed to
	// CRDGroupName
	CRDName = "configurations"

	// CRDVersion indicates what version of the CRD APIs we will be using
	CRDVersion = "v1"

	// PluginName indicates the internal configuration manager name for this plugin
	PluginName = "k8s/crd"

	// CRDForcedRefreshInterval is used to poll for any CRDs in case some were missed
	// in push-notifications
	CRDForcedRefreshInterval = 10 * time.Minute
)

// CRDShortNames indicates what shorter resource strings we can use on the CLI
var CRDShortNames = []string{
	"sbconfig",
}
