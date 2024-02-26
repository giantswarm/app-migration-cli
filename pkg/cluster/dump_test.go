package cluster

import (
	"fmt"
	"testing"

	app "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	yaml "sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestDumpAppsNames tests the migration of the app names
func TestDumpAppsNames(t *testing.T) {
	var migratedApp app.App

	appName := "loki"
	wcName := "atlastest"

	c := Cluster{
		WcName: wcName,
		SrcMC: &ManagementCluster{
			Name: "bar",
		},
		Apps: []app.App{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: "org-capa-migration-testing",
				},
				Spec: app.AppSpec{
					Name:      appName,
					Namespace: "loki",
					Version:   "0.1.0",
					Catalog:   "giantswarm",
					KubeConfig: app.AppSpecKubeConfig{
						InCluster: false,
					},
				},
			},
		},
	}

	yamlText, _ := c.migrateApps()
	yaml.Unmarshal(yamlText[0], &migratedApp)

	if migratedApp.ObjectMeta.Name != fmt.Sprintf("%s-%s", wcName, appName) {
		t.Fatalf(`App Metadata Name not correct; Is: %s; Want: %s`,
			migratedApp.ObjectMeta.Name,
			fmt.Sprintf("%s-%s", wcName, appName))
	}

	if migratedApp.Spec.Name != appName {
		t.Fatalf(`App Spec Name not correct; Is: %s; Want: %s`, migratedApp.Spec.Name, appName)
	}
}

// Test Migration with an already wc-prefixed app name
func TestDumpAlreadyPrefixedAppName(t *testing.T) {
	var migratedApp app.App

	appName := "cabbage01-service-mesh-bundle"
	wcName := "cabbage01"
	orgNamespace := "org-capa-migration-testing"

	c := Cluster{
		WcName:       wcName,
		OrgNamespace: orgNamespace,
		SrcMC: &ManagementCluster{
			Name: "bar",
		},
		Apps: []app.App{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: "org-capa-migration-testing",
				},
				Spec: app.AppSpec{
					Name:      appName,
					Namespace: "cabbage01",
					Version:   "0.1.0",
					Catalog:   "giantswarm",
					KubeConfig: app.AppSpecKubeConfig{
						InCluster: false,
					},
				},
			},
		},
	}

	yamlText, _ := c.migrateApps()
	yaml.Unmarshal(yamlText[0], &migratedApp)

	// The app name should be prefixed with the wcName
	if migratedApp.ObjectMeta.Name != appName {
		t.Fatalf(`App Metadata Name not correct; Is: %s; Want: %s`,
			migratedApp.ObjectMeta.Name,
			appName)
	}

	// The app spec name should be prefixed with the wcName
	if migratedApp.Spec.Name != appName {
		t.Fatalf(`App Spec Name not correct; Is: %s; Want: %s`, migratedApp.Spec.Name, appName)
	}
}

// Test Namespace migration
func TestDumpNamespaceMigrationOutOfCluster(t *testing.T) {
	var migratedApp app.App

	appName := "cabbage01-service-mesh-bundle"
	wcName := "cabbage01"
	appNamespace := wcName
	orgNamespace := "org-capa-migration-testing"

	c := Cluster{
		WcName:       wcName,
		OrgNamespace: orgNamespace,
		SrcMC: &ManagementCluster{
			Name: "bar",
		},
		Apps: []app.App{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: appNamespace,
				},
				Spec: app.AppSpec{
					Name:      appName,
					Namespace: appNamespace,
					Version:   "0.1.0",
					Catalog:   "giantswarm",
					KubeConfig: app.AppSpecKubeConfig{
						InCluster: false,
					},
				},
			},
		},
	}

	yamlText, _ := c.migrateApps()
	yaml.Unmarshal(yamlText[0], &migratedApp)

	// The app should be placed in the org namespace
	if migratedApp.ObjectMeta.Namespace != orgNamespace {
		t.Fatalf(`App Metadata Namespace not correct; Is: %s; Want: %s`, migratedApp.ObjectMeta.Namespace, orgNamespace)
	}

	// The app spec should not be touched
	if migratedApp.Spec.Namespace != appNamespace {
		t.Fatalf(`App Spec Namespace not correct; Is: %s; Want: %s`, migratedApp.Spec.Namespace, appNamespace)
	}

}

// Test Namespace migration
func TestDumpNamespaceMigrationInCluster(t *testing.T) {
	var migratedApp app.App

	appName := "cabbage01-service-mesh-bundle"
	wcName := "cabbage01"
	appNamespace := wcName
	orgNamespace := "org-capa-migration-testing"

	c := Cluster{
		WcName:       wcName,
		OrgNamespace: orgNamespace,
		SrcMC: &ManagementCluster{
			Name: "bar",
		},
		Apps: []app.App{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: appNamespace,
				},
				Spec: app.AppSpec{
					Name:      appName,
					Namespace: appNamespace,
					Version:   "0.1.0",
					Catalog:   "giantswarm",
					KubeConfig: app.AppSpecKubeConfig{
						InCluster: true,
					},
				},
			},
		},
	}

	yamlText, _ := c.migrateApps()
	yaml.Unmarshal(yamlText[0], &migratedApp)

	// The app should be placed in the org namespace
	if migratedApp.ObjectMeta.Namespace != orgNamespace {
		t.Fatalf(`App Metadata Namespace not correct; Is: %s; Want: %s`, migratedApp.ObjectMeta.Namespace, orgNamespace)
	}

	// The app spec should be placed in the org namespace
	if migratedApp.Spec.Namespace != orgNamespace {
		t.Fatalf(`App Spec Namespace not correct; Is: %s; Want: %s`, migratedApp.Spec.Namespace, orgNamespace)
	}
}

// Test UserConfig migration
func TestDumpUserConfigmapMigration(t *testing.T) {
	var migratedApp app.App
	var migratedCm corev1.ConfigMap

	appName := "cabbage01-service-mesh-bundle"
	cmName := "foobar"
	wcName := "cabbage01"
	orgNamespace := "org-capa-migration-testing"

	// Create a fake configmap and add it to the fake client
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: wcName,
		},
	}
	var client client.Client
	initObjs := []runtime.Object{cm}
	client = fake.NewFakeClient(initObjs...)

	c := Cluster{
		WcName:       wcName,
		OrgNamespace: orgNamespace,
		SrcMC: &ManagementCluster{
			Name:             "bar",
			KubernetesClient: client,
		},
		Apps: []app.App{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: "org-capa-migration-testing",
				},
				Spec: app.AppSpec{
					Name:      appName,
					Namespace: "cabbage01",
					Version:   "0.1.0",
					Catalog:   "giantswarm",
					KubeConfig: app.AppSpecKubeConfig{
						InCluster: false,
					},
					UserConfig: app.AppSpecUserConfig{
						ConfigMap: app.AppSpecUserConfigConfigMap{
							Name:      cmName,
							Namespace: wcName,
						},
					},
				},
			},
		},
	}

	yamlText, _ := c.migrateApps()
	yaml.Unmarshal(yamlText[0], &migratedCm)
	yaml.Unmarshal(yamlText[1], &migratedApp)

	// The cm should be referrenced from the org namespace
	if migratedApp.Spec.UserConfig.ConfigMap.Namespace != orgNamespace {
		t.Fatalf(`App UserConfig ConfigMap Namespace not correct; Is: %s; Want: %s`, migratedApp.Spec.UserConfig.ConfigMap.Namespace, orgNamespace)
	}

	// The cm should be placed in the org namespace
	if migratedCm.ObjectMeta.Namespace != orgNamespace {
		t.Fatalf(`ConfigMap of UserConfig Namespace not correct; Is: %s; Want: %s`, migratedCm.ObjectMeta.Namespace, orgNamespace)
	}

	// The cm should be referrenced with the new name
	if migratedApp.Spec.UserConfig.ConfigMap.Name != fmt.Sprintf("%s-%s", wcName, cmName) {
		t.Fatalf(`App UserConfig ConfigMap Name not correct; Is: %s; Want: %s`, migratedApp.Spec.UserConfig.ConfigMap.Name, fmt.Sprintf("%s-%s", wcName, cmName))
	}

	// The cm should be renamed
	if migratedCm.ObjectMeta.Name != fmt.Sprintf("%s-%s", wcName, cmName) {
		t.Fatalf(`ConfigMap of UserConfig Namespace not correct; Is: %s; Want: %s`, migratedCm.ObjectMeta.Name, fmt.Sprintf("%s-%s", wcName, cmName))
	}

}

// Test ExtraConfig migration
func TestExtraConfigSecretMigration(t *testing.T) {
	var migratedApp app.App
	var migratedSecret corev1.Secret

	appName := "cabbage01-service-mesh-bundle"
	secretName := "foobar"
	wcName := "cabbage01"
	orgNamespace := "org-capa-migration-testing"

	// Create a fake secret and add it to the fake client
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: wcName,
		},
	}

	var client client.Client
	initObjs := []runtime.Object{secret}
	client = fake.NewFakeClient(initObjs...)

	c := Cluster{
		WcName:       wcName,
		OrgNamespace: orgNamespace,
		SrcMC: &ManagementCluster{
			Name:             "bar",
			KubernetesClient: client,
		},
		Apps: []app.App{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: "org-capa-migration-testing",
				},
				Spec: app.AppSpec{
					Name:      appName,
					Namespace: "cabbage01",
					Version:   "0.1.0",
					Catalog:   "giantswarm",
					KubeConfig: app.AppSpecKubeConfig{
						InCluster: false,
					},
					ExtraConfigs: []app.AppExtraConfig{
						{
							Kind:      "secret",
							Name:      secretName,
							Namespace: wcName,
							Priority:  64,
						},
					},
				},
			},
		},
	}

	yamlText, _ := c.migrateApps()
	yaml.Unmarshal(yamlText[0], &migratedSecret)
	yaml.Unmarshal(yamlText[1], &migratedApp)

	// ExtraConfig Priority should be preserved
	if migratedApp.Spec.ExtraConfigs[0].Priority != 64 {
		t.Fatalf(`App ExtraConfig Priority not correct; Is: %d; Want: %d`, migratedApp.Spec.ExtraConfigs[0].Priority, 64)
	}

	// The secret should be referrenced from the org namespace
	if migratedApp.Spec.ExtraConfigs[0].Namespace != orgNamespace {
		t.Fatalf(`App ExtraConfig Secret Namespace not correct; Is: %s; Want: %s`, migratedApp.Spec.ExtraConfigs[0].Namespace, orgNamespace)
	}

	// The secret should be placed in the org namespace
	if migratedSecret.ObjectMeta.Namespace != orgNamespace {
		t.Fatalf(`Secret of ExtraConfig Namespace not correct; Is: %s; Want: %s`, migratedSecret.ObjectMeta.Namespace, orgNamespace)
	}

	// The secret should be referrenced with the new name
	if migratedApp.Spec.ExtraConfigs[0].Name != fmt.Sprintf("%s-%s", wcName, secretName) {
		t.Fatalf(`App ExtraConfig Secret Name not correct; Is: %s; Want: %s`, migratedApp.Spec.ExtraConfigs[0].Name, fmt.Sprintf("%s-%s", wcName, secretName))
	}

	// The secret should be renamed
	if migratedSecret.ObjectMeta.Name != fmt.Sprintf("%s-%s", wcName, secretName) {
		t.Fatalf(`Secret of ExtraConfig Namespace not correct; Is: %s; Want: %s`, migratedSecret.ObjectMeta.Name, fmt.Sprintf("%s-%s", wcName, secretName))
	}

}
