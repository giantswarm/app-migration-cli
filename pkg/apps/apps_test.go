package apps

import (
	"errors"
	"testing"

	app "github.com/giantswarm/apiextensions-application/api/v1alpha1"

)

func TestFilterAppCRsEmptyReturn(t *testing.T) {
  emptyApp := []app.App{}

  _, err := filterAppCRs(emptyApp)

  if ! errors.Is(err, EmptyAppsError) {
    t.Fatalf("Empty App List is not returning error")
  }
}

func TestFilteringOfBundledApps(t *testing.T) {
  appLabels := map[string]string{}
  appLabels["giantswarm.io/managed-by"] = ""

  newApp := app.App{}

  newApp.Labels = appLabels

  appList := []app.App{
    newApp,
  }

  _, err := filterAppCRs(appList)
  if ! errors.Is(err, EmptyAppsError) {
    t.Fatalf("App Bundle should be filtered for migration")
  }
}
 
func TestFilteringOfDefaultApps(t *testing.T) {
  newApp := app.App{}
  newApp.Spec.Catalog = "default"

  appList := []app.App{
    newApp,
  }

  _, err := filterAppCRs(appList)
  if ! errors.Is(err, EmptyAppsError) {
    t.Fatalf("Apps from the `default` catalog should be filtered")
  }
}
