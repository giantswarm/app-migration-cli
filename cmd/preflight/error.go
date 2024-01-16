package preflight

import (
  "github.com/giantswarm/microerror"
)

var invalidFlagsError = &microerror.Error{
  Kind: "invalidFlagsError",
}

var invalidConfigError = &microerror.Error{
  Kind: "invalidConfigError",
}

var migrationBlocked = &microerror.Error{
  Kind: "migrationBlocked",
}
