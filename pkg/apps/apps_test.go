package apps

import (
	"errors"
	"testing"

	app "github.com/giantswarm/apiextensions-application/api/v1alpha1"
)

func TestFilterAppCRsEmptyReturn(t *testing.T) {
	emptyApp := []app.App{}

	_, err := filterAppCRs(emptyApp)

	if !errors.Is(err, EmptyAppsError) {
		t.Fatalf("Empty App List is not returning error")
	}
}

func TestFilteringOfBundledApps(t *testing.T) {
	appLabels := map[string]string{}
	appLabels["giantswarm.io/managed-by"] = "testc-observavility-bundle"

	newApp := app.App{}

	newApp.Labels = appLabels

	appList := []app.App{
		newApp,
	}

	_, err := filterAppCRs(appList)
	if !errors.Is(err, EmptyAppsError) {
		t.Fatalf("App Bundle should be filtered for migration")
	}
}

func TestFilteringOfBundledSecurityApps(t *testing.T) {
	appLabels := map[string]string{}
	appLabels["app.kubernetes.io/name"] = "security-bundle"
	appLabels["giantswarm.io/managed-by"] = "cluster-operator"

	newApp := app.App{}

	newApp.Labels = appLabels

	appList := []app.App{
		newApp,
	}

	_, err := filterAppCRs(appList)
	if !errors.Is(err, EmptyAppsError) {
		t.Fatalf("App Bundle should be filtered for migration")
	}
}

func TestNoFilteringOfCustomerApps(t *testing.T) {
	appLabels := map[string]string{}
	appLabels["giantswarm.io/managed-by"] = "customer"

	newApp := app.App{}

	newApp.Labels = appLabels

	appList := []app.App{
		newApp,
	}

	_, err := filterAppCRs(appList)
	if err != nil && !errors.Is(err, EmptyAppsError) {
		t.Fatalf("App Bundle should be filtered for migration")
	}
}

func TestFilteringOfAppOperatorApps(t *testing.T) {
	appLabels := map[string]string{}
	appLabels["giantswarm.io/managed-by"] = "app-operator"

	newApp := app.App{}

	newApp.Labels = appLabels

	appList := []app.App{
		newApp,
	}

	_, err := filterAppCRs(appList)
	if !errors.Is(err, EmptyAppsError) {
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
	if !errors.Is(err, EmptyAppsError) {
		t.Fatalf("Apps from the `default` catalog should be filtered")
	}
}

func TestFilteringOfNamedApps(t *testing.T) {
	newApp := app.App{}
	newApp.Spec.Name = "k8s-initiator-app"

	appList := []app.App{
		newApp,
	}

	_, err := filterAppCRs(appList)
	if !errors.Is(err, EmptyAppsError) {
		t.Fatalf("Apps named `%s` catalog should be filtered", newApp.Spec.Name)
	}
}
