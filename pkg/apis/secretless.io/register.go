package secretless

import "os"

var (
	// GroupName is the CRD TLD identifier
	GroupName = "secretless" + os.Getenv("SECRETLESS_CRD_SUFFIX") + ".io"
)
