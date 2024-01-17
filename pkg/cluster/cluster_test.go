package cluster

import(
  "fmt"
  "os"
  "testing"
)

func TestAppYamlFileWithOverride(t *testing.T) {
  c := Cluster{
    WcName: "foo",
    SrcMC: &ManagementCluster{
      Name: "bar",
    },
  }

  res := c.AppYamlFile("test.yaml")
  dir, _ := os.Getwd()
  want := fmt.Sprintf("%s/test.yaml", dir)

  if want != res {
    t.Fatalf(`Result: %s, want %s`, res, want)
  }
}

func TestAppYamlFileWithEmpty(t *testing.T) {
  c := Cluster{
    WcName: "foo",
    SrcMC: &ManagementCluster{
      Name: "bar",
    },
  }

  res := c.AppYamlFile("")
  dir, _ := os.Getwd()
  want := fmt.Sprintf("%s/%s-%s-apps.yaml", dir, c.SrcMC.Name, c.WcName)

  if want != res {
    t.Fatalf(`Result: %s, want %s`, res, want)
  }
}
