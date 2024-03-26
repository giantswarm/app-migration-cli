package apps

import (
	"context"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	app "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
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

	filteredApps, err := filterAppCRs(objList.Items)

	return filteredApps, err
}

// blacklist certain apps for migration
func filterAppCRs(allApps []app.App) ([]app.App, error) {
	var filteredApps []app.App
appLoop:
	for _, application := range allApps {
		// skip "default" apps; these should be installed by default on the MC
		if application.Spec.Catalog == "default" {
			continue
		}

		// skip bundled apps as we only migrate their parent
		// todo: verify thats formally correct
		labels := application.GetLabels()
		for key := range labels {
			if strings.Contains(key, "giantswarm.io/managed-by") {
				// we skip this app completly
				continue appLoop
			}
		}

		// skip specific apps that are no longer supported on CAPI
		if application.Spec.Name == "k8s-initiator-app" {
			continue
		}

		filteredApps = append(filteredApps, application)
	}

	if len(filteredApps) == 0 {
		return nil, microerror.Maskf(EmptyAppsError, "No non-default apps found for migration")
	}

	return filteredApps, nil
}
