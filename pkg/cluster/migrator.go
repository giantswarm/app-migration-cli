package cluster

import (
	"os"
	"strings"
	"fmt"

	"golang.org/x/net/context"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"
  k8syaml "sigs.k8s.io/yaml"

  app "github.com/giantswarm/kubectl-gs/v2/pkg/template/app"
//  apps "github.com/giantswarm/apiextensions-application/api/v1alpha1"
  corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"

)

func (c *Cluster) DumpApps() error {
  
  var numberOfAppsToMigrate int

  // we write the apps to a yaml-file, which gets applied later
	f, err := os.OpenFile(nonDefaultAppYamlFile(c.SrcMC.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)

	if err != nil {
		return microerror.Mask(err)
	}

  appLoop:
    for _,application := range c.Apps {
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

    numberOfAppsToMigrate += 1

// 	DefaultingEnabled          bool
// 	ExtraLabels                map[string]string
// 	ExtraAnnotations           map[string]string
// 	UseClusterValuesConfig     bool

    // todo: app operator version; does it impact the migration?
    // todo: how to deal with ExtraLabels and Extrannotations?
    newApp := app.Config{
      AppName:      application.ObjectMeta.Name,
			Catalog:      application.Spec.Catalog,
      Cluster:      c.WcName,
      InCluster:    application.Spec.KubeConfig.InCluster,
			Name:         application.Spec.Name,
			Namespace:    application.Spec.Namespace,
			Version:      application.Spec.Version,
      ExtraLabels:  application.GetLabels(),
      ExtraAnnotations: application.GetAnnotations(),
      Organization: organizationFromNamespace(c.OrgNamespace),
		}

    fmt.Printf("org: %s", newApp.Namespace)

    // make sure we trim the clustername if it somehow was prefixed on the app
    metadataName := strings.TrimLeft(application.GetName(), c.WcName)
    // now prefix our app with the cluster
    newApp.AppName = fmt.Sprintf("%s-%s", c.WcName, metadataName)

    // todo: secret missing?
    if application.Spec.Config.ConfigMap.Name == fmt.Sprintf("%s-cluster-values", c.WcName) {
      newApp.UseClusterValuesConfig = true
    }

    if application.Spec.ExtraConfigs != nil {

      for _, extraConfig := range application.Spec.ExtraConfigs {
        obj, err := migrateAppConfigObject(
            c.SrcMC.KubernetesClient,
            strings.ToLower(extraConfig.Kind),
            c.WcName,
            extraConfig.Name, 
            extraConfig.Namespace, 
            newApp.Organization)
        if err != nil {
          return microerror.Mask(err)
        }

        newApp.ExtraConfigs = append(newApp.ExtraConfigs, applicationv1alpha1.AppExtraConfig{
          Kind: obj.Kind,
          Name: obj.Name,
          Namespace: obj.Namespace,
          Priority: extraConfig.Priority,
        })

        if _, err := f.Write([]byte(fmt.Sprintf("%s---\n", obj.Yaml))); err != nil {
          return microerror.Mask(err)
        }
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

    // apps on the WC should go to the org namespace
    if application.Spec.KubeConfig.InCluster == false {
      newApp.Organization = organizationFromNamespace(c.SrcMC.Namespace)
    }

    if application.Spec.UserConfig.ConfigMap.Name != "" {

      configmap, err := migrateAppConfigObject(
          c.SrcMC.KubernetesClient,
          "configmap",
          c.WcName,
          application.Spec.UserConfig.ConfigMap.Name,
          application.Spec.UserConfig.ConfigMap.Namespace,
          newApp.Organization)
      if err != nil {
        return microerror.Mask(err)
      }

      newApp.UserConfigConfigMapName = configmap.Name

      if _, err := f.Write([]byte(fmt.Sprintf("%s---\n", configmap.Yaml))); err != nil {
        return microerror.Mask(err)
      }
    }

    if application.Spec.UserConfig.Secret.Name != "" {
      newApp.UserConfigSecretName = application.Spec.UserConfig.Secret.Name

      secret, err := migrateAppConfigObject(
        c.SrcMC.KubernetesClient,
        "secret",
        c.WcName,
        application.Spec.UserConfig.Secret.Name,
        application.Spec.UserConfig.Secret.Namespace,
        newApp.Organization)
      if err != nil {
        return microerror.Mask(err)
      }

      newApp.UserConfigSecretName = secret.Name

      if _, err := f.Write([]byte(fmt.Sprintf("%s---\n", secret.Yaml))); err != nil {
        return microerror.Mask(err)
      }
    }

    appYAML, err := app.NewAppCR(newApp)
    if err != nil {
      return microerror.Mask(err)
    }

    if _, err := f.Write([]byte(fmt.Sprintf("%s---\n", appYAML))); err != nil {
      return microerror.Mask(err)
    }

  }
	
  fmt.Printf("Scheduled %d non-default apps for migration\n", numberOfAppsToMigrate)

  if err := f.Close(); err != nil {
    return microerror.Mask(err)
  }

  return nil
}

func nonDefaultAppYamlFile(clusterName string) string {
	wd, _ := os.Getwd()
	return fmt.Sprintf("%s/%s-apps.yaml", wd, clusterName)
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
    name := strings.TrimLeft(resourceName, clusterName)
    config.Name = fmt.Sprintf("%s-%s", clusterName, name)


  switch resourceKind {
    case "secret":
      var secret corev1.Secret
      config.Kind = "secret"

      err := k8sClient.Get(context.TODO(), client.ObjectKey{
        Name: resourceName,
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
          Name: config.Name,
          Namespace: config.Namespace,
          Labels: secret.GetLabels(),
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
          Name: resourceName,
          Namespace: namespace,
        }, &cm)

      if err != nil {
        return AppExtraConfig{},microerror.Mask(err)
      }

      newCm := &corev1.ConfigMap{
        TypeMeta: metav1.TypeMeta{
          Kind:       "ConfigMap",
          APIVersion: "v1",
        },
        ObjectMeta: metav1.ObjectMeta{
          Name: config.Name,
          Namespace: config.Namespace,
          Labels: cm.GetLabels(),
          Annotations: cm.GetAnnotations(),
        },
        Data: cm.Data,
      }

      config.Yaml, err = k8syaml.Marshal(newCm)
      if err != nil {
        return AppExtraConfig{}, microerror.Mask(err)
      }

    default:
      return AppExtraConfig{},fmt.Errorf("unsupported resource kind: %s", resourceKind)
    }

    return config, nil
}
