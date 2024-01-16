package apps

import (
  "github.com/giantswarm/microerror"
)

var emptyAppsError = &microerror.Error{
  Kind: "emptyAppsError",
}
