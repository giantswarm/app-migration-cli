package apply

import (
	"github.com/giantswarm/microerror"
)

// Flags represents all the flags that can be set via the command line
type Flags struct {
	sourceFile  string
	dstMC  string
	srcMC  string
  wcName string
  finalizer bool
  orgNamespace string
}

func (f *Flags) Validate() error {
	if f.sourceFile == "" {
		return microerror.Maskf(invalidFlagsError, "Filename must not be empty")
	}

  if f.srcMC == "" {
		return microerror.Maskf(invalidFlagsError, "SourceMC must not be empty")
	}

	if f.dstMC == "" {
		return microerror.Maskf(invalidFlagsError, "DestinationMC must not be empty")
	}

	if f.wcName == "" {
		return microerror.Maskf(invalidFlagsError, "WorkloadClusterName must not be empty")
	}

  if f.orgNamespace == "" {
		return microerror.Maskf(invalidFlagsError, "OrgNamespace must not be empty")
	}

	return nil
}
