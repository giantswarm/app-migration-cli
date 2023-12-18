package cluster

import (
	"github.com/giantswarm/microerror"
)

var clusterNotFound = &microerror.Error{
	Kind: "clusterNotFound",
}

var clusterNameNotFound = &microerror.Error{
	Kind: "clusterNameNotFound",
}

var clusterUnhealthy = &microerror.Error{
	Kind: "clusterUnhealthy",
}


