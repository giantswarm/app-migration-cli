package cluster

import (
	"fmt"
	"os"
	"strings"

	"github.com/giantswarm/microerror"
	"golang.org/x/net/context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8syaml "sigs.k8s.io/yaml"

	app "github.com/giantswarm/kubectl-gs/v2/pkg/template/app"
	//  apps "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Cluster) DumpApps(f *os.File) error {

	yaml, err := c.migrateApps()
	if err != nil {
		return microerror.Mask(err)
	}

	for _, obj := range yaml {
		if _, err := f.Write([]byte(fmt.Sprintf("%s---\n", obj))); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (c *Cluster) shouldSkipConfigMapOrSecretMigration(configMapOrSecretName string) bool {
	// Do not copy cluster values config map / secret to CAPI MC. A new one
	// on the CAPI MC should automatically be created by cluster-apps-operator,
	// using new values (e.g. different `baseDomain` from vintage).
	// Copying it would result in cluster-apps-operator not reconciling it.
	// In the `app-migration-cli apply` subcommand, we even wait for the new
	// config map / secret to be available in order to allow Apps to deploy correctly.
	return configMapOrSecretName == fmt.Sprintf("%s-cluster-values", c.WcName)
}

func (c *Cluster) migrateApps() ([][]byte, error) {

	var yaml [][]byte

	for _, application := range c.Apps {
		// 	DefaultingEnabled          bool
		// 	UseClusterValuesConfig     bool

		// todo: app operator version; does it impact the migration?
		// todo: how to deal with ExtraLabels and Extrannotations?
		newApp := app.Config{
			AppName:          application.ObjectMeta.Name,
			Catalog:          application.Spec.Catalog,
			Cluster:          c.WcName,
			InCluster:        application.Spec.KubeConfig.InCluster,
			Name:             application.Spec.Name,
			Namespace:        application.Spec.Namespace,
			Version:          application.Spec.Version,
			ExtraLabels:      application.GetLabels(),
			ExtraAnnotations: application.GetAnnotations(),
			Organization:     organizationFromNamespace(c.OrgNamespace),
		}

		// make sure we trim the clustername if it somehow was prefixed on the app
		appName := strings.TrimPrefix(application.GetName(), c.WcName)
		// in case we trimmed the clustername, we might need to trim the trailing dash
		// now as well.
		appName = strings.TrimPrefix(appName, "-")
		// now prefix the app with the wcName
		newApp.AppName = fmt.Sprintf("%s-%s", c.WcName, appName)

		// apps on the MC should go to the org namespace
		if application.Spec.KubeConfig.InCluster {
			newApp.Namespace = c.OrgNamespace
		}

		if application.Spec.Config.ConfigMap.Name == fmt.Sprintf("%s-cluster-values", c.WcName) {
			newApp.UseClusterValuesConfig = true
		}

		if application.Spec.ExtraConfigs != nil {
			for _, extraConfig := range application.Spec.ExtraConfigs {
				if (strings.ToLower(extraConfig.Kind) == "configmap" || strings.ToLower(extraConfig.Kind) == "secret") && c.shouldSkipConfigMapOrSecretMigration(extraConfig.Name) {
					continue
				}

				obj, err := migrateAppConfigObject(
					c.SrcMC.KubernetesClient,
					strings.ToLower(extraConfig.Kind),
					c.WcName,
					extraConfig.Name,
					extraConfig.Namespace,
					newApp.Organization)

				if err != nil {
					return nil, microerror.Mask(err)
				}

				newApp.ExtraConfigs = append(newApp.ExtraConfigs, applicationv1alpha1.AppExtraConfig{
					Kind:      obj.Kind,
					Name:      obj.Name,
					Namespace: obj.Namespace,
					Priority:  extraConfig.Priority,
				})

				yaml = append(yaml, obj.Yaml)
			}
		}

		if application.Spec.CatalogNamespace != "" {
			newApp.CatalogNamespace = application.Spec.CatalogNamespace
		}

		if application.Spec.NamespaceConfig.Labels != nil {
			newApp.NamespaceConfigLabels = application.Spec.NamespaceConfig.Labels
		}
		if application.Spec.NamespaceConfig.Annotations != nil {
			newApp.NamespaceConfigAnnotations = application.Spec.NamespaceConfig.Annotations
		}

		if application.Spec.Install.Timeout != nil {
			newApp.InstallTimeout = application.Spec.Install.Timeout
		}
		if application.Spec.Rollback.Timeout != nil {
			newApp.RollbackTimeout = application.Spec.Rollback.Timeout
		}
		if application.Spec.Uninstall.Timeout != nil {
			newApp.UninstallTimeout = application.Spec.Uninstall.Timeout
		}
		if application.Spec.Upgrade.Timeout != nil {
			newApp.UpgradeTimeout = application.Spec.Upgrade.Timeout
		}

		if application.Spec.UserConfig.ConfigMap.Name != "" && !c.shouldSkipConfigMapOrSecretMigration(application.Spec.UserConfig.ConfigMap.Name) {
			configmap, err := migrateAppConfigObject(
				c.SrcMC.KubernetesClient,
				"configmap",
				c.WcName,
				application.Spec.UserConfig.ConfigMap.Name,
				application.Spec.UserConfig.ConfigMap.Namespace,
				newApp.Organization)

			if err != nil {
				return nil, microerror.Mask(err)
			}

			newApp.UserConfigConfigMapName = configmap.Name

			yaml = append(yaml, configmap.Yaml)
		}

		if application.Spec.UserConfig.Secret.Name != "" && !c.shouldSkipConfigMapOrSecretMigration(application.Spec.UserConfig.Secret.Name) {
			newApp.UserConfigSecretName = application.Spec.UserConfig.Secret.Name

			secret, err := migrateAppConfigObject(
				c.SrcMC.KubernetesClient,
				"secret",
				c.WcName,
				application.Spec.UserConfig.Secret.Name,
				application.Spec.UserConfig.Secret.Namespace,
				newApp.Organization)

			if err != nil {
				return nil, microerror.Mask(err)
			}

			newApp.UserConfigSecretName = secret.Name

			yaml = append(yaml, secret.Yaml)
		}

		appYAML, err := app.NewAppCR(newApp)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		yaml = append(yaml, appYAML)
	}

	return yaml, nil
}

func organizationFromNamespace(namespace string) string {
	return strings.TrimPrefix(namespace, "org-")
}

func migrateAppConfigObject(k8sClient client.Client, resourceKind string, clusterName string, resourceName string, namespace string, organization string) (AppExtraConfig, error) {

	var config AppExtraConfig

	if namespace != "default" && namespace != "giantswarm" {
		// we force-migrate the objects to the org-namespace
		// if they were in a custom one before

		config.Namespace = fmt.Sprintf("%s-%s", "org", organization)
	} else {
		config.Namespace = namespace
	}

	// make sure we trim the clustername if it somehow was prefixed on the app
	name := strings.TrimPrefix(resourceName, clusterName)
	name = strings.TrimPrefix(name, "-")
	config.Name = fmt.Sprintf("%s-%s", clusterName, name)

	switch resourceKind {
	case "secret":
		var secret corev1.Secret
		config.Kind = "secret"

		err := k8sClient.Get(context.TODO(), client.ObjectKey{
			Name:      resourceName,
			Namespace: namespace,
		}, &secret)

		if err != nil {
			return AppExtraConfig{}, microerror.Mask(err)
		}

		newSecret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        config.Name,
				Namespace:   config.Namespace,
				Labels:      secret.GetLabels(),
				Annotations: secret.GetAnnotations(),
			},
			Data: secret.Data,
		}
		config.Yaml, err = k8syaml.Marshal(newSecret)

		if err != nil {
			return AppExtraConfig{}, microerror.Mask(err)
		}

	case "configmap":
		var cm corev1.ConfigMap
		config.Kind = "configMap"

		err := k8sClient.Get(context.TODO(), client.ObjectKey{
			Name:      resourceName,
			Namespace: namespace,
		}, &cm)

		if err != nil {
			return AppExtraConfig{}, microerror.Mask(err)
		}

		newCm := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        config.Name,
				Namespace:   config.Namespace,
				Labels:      cm.GetLabels(),
				Annotations: cm.GetAnnotations(),
			},
			Data: cm.Data,
		}

		config.Yaml, err = k8syaml.Marshal(newCm)
		if err != nil {
			return AppExtraConfig{}, microerror.Mask(err)
		}

	default:
		return AppExtraConfig{}, fmt.Errorf("unsupported resource kind: %s", resourceKind)
	}

	return config, nil
}
