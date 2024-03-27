package apps

import (
	"github.com/giantswarm/microerror"
)

var EmptyAppsError = &microerror.Error{
	Kind: "emptyAppsError",
}
