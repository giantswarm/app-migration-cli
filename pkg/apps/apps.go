package apps

import (
  "context"

  "sigs.k8s.io/controller-runtime/pkg/client"


  "github.com/giantswarm/microerror"
  app "github.com/giantswarm/apiextensions-application/api/v1alpha1"
)

func GetAppCRs(k8sClient client.Client, clusterName string) ([]app.App, error) {
  objList := &app.AppList{}

  // todo: not possible to filter on "spec.catalog" bc/ cached list not indexed?
  selector := client.MatchingFields{"metadata.namespace": clusterName}
  //selector := client.MatchingLabels{"app.kubernetes.io/name"
  err := k8sClient.List(context.TODO(), objList, selector)
  if err != nil {
    return nil, microerror.Mask(err)
  }

  if len(objList.Items) == 0 {
    return nil, microerror.Maskf(emptyAppsError, "No non-default apps found for migration")
  }

  return objList.Items, nil
}
