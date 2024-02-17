package cluster

import (
	//  "fmt"
	"fmt"
	"testing"

	app "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	yaml "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MyApp struct {
  ObjectMeta metav1.ObjectMeta `yaml:"metadata"`
  Kind string `yaml:"kind"`
  Spec app.AppSpec `yaml:"spec"`
}

func TestDumpAppsNames(t *testing.T) {
  var retApp MyApp
  var apps []app.App

  appName := "loki"
  wcName := "atlastest"

  app := app.App{
    ObjectMeta: metav1.ObjectMeta{
      Name: appName,
      Namespace: "org-capa-migration-testing",
    },
    Spec: app.AppSpec{
      Name: appName,
      Namespace: "loki",
      Version: "0.1.0",
      Catalog: "giantswarm",
      KubeConfig: app.AppSpecKubeConfig{
        InCluster: false,
      },
    },
  }

  c := Cluster{
    WcName: wcName,
    SrcMC: &ManagementCluster{
      Name: "bar",
    },
    Apps: append(apps, app),
  }

  yamlText, _ := c.migrateApps()
  yaml.Unmarshal(yamlText, &retApp)

  if retApp.ObjectMeta.Name != fmt.Sprintf("%s-%s", wcName, appName) {
    t.Fatalf(`App Metadata Name not correct; Is: %s; Want: %s`,
        retApp.ObjectMeta.Name,
        fmt.Sprintf("%s-%s", wcName, appName))
  }

  if retApp.Spec.Name != appName {
    t.Fatalf(`App Spec Name not correct; Is: %s; Want: %s`, retApp.Spec.Name, appName)
  }
 }


