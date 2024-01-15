package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
  "slices"

	"github.com/giantswarm/microerror"
	"golang.org/x/net/context"

//  "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "k8s.io/api/core/v1"

	gsv1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	chart "github.com/giantswarm/apiextensions-application/api/v1alpha1"
  apps "github.com/giantswarm/apiextensions-application/api/v1alpha1"

)

const finalizer string = "giantswarm.io/app-migration-cli"

var (
  scheme = runtime.NewScheme()

  // wc in any of these "conditon" are considered healthy
  // check `kubectl gs get cluster <name>`
  validClusterStates = []string{
    "Created",
    "Updating",
    "Updated",
  }
)

func init() {
  _ = clientgoscheme.AddToScheme(scheme)
  _ = capi.AddToScheme(scheme)
 	_ = gsv1alpha3.AddToScheme(scheme)
  //	_ = kubeadmv1beta1.AddToScheme(scheme)
  _ = chart.AddToScheme(scheme)
}

type Cluster struct {
  WcName string
  Apps []apps.App

  SrcMC *ManagementCluster
  DstMC *ManagementCluster
}

type ManagementCluster struct {
  Name      string
  Namespace string

  KubernetesClient client.Client
}

type AppExtraConfig struct {
  Kind      string
  Name      string
  Namespace string
  Yaml      []byte
}

func (c *ManagementCluster) GetWCHealth(clusterName string) (string, error) {

  ctx := context.TODO()

  capiCluster, err := c.getCluster(ctx, clusterName)
  if err != nil {
    return "", microerror.Mask(err)
  }

  var awsCluster *gsv1alpha3.AWSCluster

  if capiCluster.Spec.InfrastructureRef.Name != "" {
    awsCluster, err = c.getAwsClusterByName(ctx, capiCluster.Spec.InfrastructureRef.Name)
    if err != nil {
      return "", microerror.Mask(err)
    }
  } else {
    return "", microerror.Maskf(clusterNameNotFound, "AWSCluster Name not found for %s", clusterName)
  }

  health := getLastAwsCondition(awsCluster.Status.Cluster.Conditions)
  if slices.Contains(validClusterStates, health) {
    return health, nil
  } else {
    return "", microerror.Maskf(clusterUnhealthy, "WorkloadCluster not in a healthy condition")
  }

}

func getLastAwsCondition(cond []gsv1alpha3.CommonClusterStatusCondition) string {
  if len(cond) < 1 {
    return "n/a"
  }

  return cond[0].Condition
}

func (c *ManagementCluster) getAwsClusterByName(ctx context.Context, clusterName string) (*gsv1alpha3.AWSCluster, error) {
  objList := &gsv1alpha3.AWSClusterList{}
  selector := client.MatchingLabels{capi.ClusterNameLabel: clusterName}

  err := c.KubernetesClient.List(ctx, objList, selector)
  if err != nil {
    return nil, microerror.Mask(err)
  }

  if len(objList.Items) == 0 {
    return nil, microerror.Maskf(clusterNotFound, "Cluster not found for %s", clusterName)
  }

  if len(objList.Items) > 1 {
    return nil, microerror.Maskf(clusterNotFound, "More than one Cluster found with name %s", clusterName)
  }


  return &objList.Items[0], nil
}

func (c *ManagementCluster) getCluster(ctx context.Context, clusterName string) (*capi.Cluster, error) {
  objList := &capi.ClusterList{}
  selector := client.MatchingLabels{capi.ClusterNameLabel: clusterName}
  err := c.KubernetesClient.List(ctx, objList, selector)
  if err != nil {
    return nil, microerror.Mask(err)
  }

  if len(objList.Items) == 0 {
    return nil, microerror.Maskf(clusterNotFound, "Cluster not found for %s", clusterName)
  }

  if len(objList.Items) > 1 {
    return nil, microerror.Maskf(clusterNotFound, "More than one Cluster found with name %s", clusterName)
  }

  return &objList.Items[0], nil
}

// todo: migrate to single cluster login
func Login(srcMC string, dstMc string) (*Cluster, error) {
  srcMcClient, _, err := loginOrReuseKubeconfig([]string{srcMC})
  if err != nil {
    return nil, microerror.Mask(err)
  }
  dstMcClient, _, err := loginOrReuseKubeconfig([]string{dstMc})
  if err != nil {
    return nil, microerror.Mask(err)
  }

  return &Cluster {
    SrcMC:   &ManagementCluster{
      Name: srcMC,
      KubernetesClient: srcMcClient,
    },
    DstMC: & ManagementCluster{
      Name: dstMc,
      KubernetesClient: dstMcClient,
    },
  }, nil
}

// LoginOrReuseKubeconfig will return k8s client for the specific wc or MC client, it will try if there is already existing context or login if its missing
func loginOrReuseKubeconfig(cluster []string) (client.Client, kubernetes.Interface, error) {
  ctrlClient, clientSet, err := getK8sClientFromKubeconfig(contextNameFromCluster(cluster))
  if err != nil && strings.Contains(err.Error(), "does not exist") {
    // login
    fmt.Printf("Context for cluster %s not found, executing 'opsctl login', check your browser window.\n", cluster)
    err = loginIntoCLuster(cluster)
    if err != nil {
      return nil, nil, microerror.Mask(err)
    }
    // now retry
    ctrlClient, clientSet, err = getK8sClientFromKubeconfig(contextNameFromCluster(cluster))
    if err != nil {
      return nil, nil, microerror.Mask(err)
    }
  } else if err != nil {
    return nil, nil, microerror.Mask(err)
  }
  return ctrlClient, clientSet, nil
}

func getK8sClientFromKubeconfig(contextName string) (client.Client, kubernetes.Interface, error) {
  kubeconfigFile := os.Getenv("KUBECONFIG")
  if kubeconfigFile == "" {
    home, err := os.UserHomeDir()
    if err != nil {
      return nil, nil, microerror.Mask(err)
    }
    kubeconfigFile = fmt.Sprintf("%s/.kube/config", home)
  }

  config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
    &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigFile},
    &clientcmd.ConfigOverrides{
      CurrentContext: contextName,
    }).ClientConfig()

  if err != nil {
    return nil, nil, microerror.Mask(err)
  }

  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, nil, microerror.Mask(err)
  }
  v, err := clientset.ServerVersion()
  if err != nil {
    return nil, nil, microerror.Mask(err)
  }
  fmt.Printf("Connected to %s, k8s server version %s\n", contextName, v.String())

  ctrlClient, err := client.New(config, client.Options{Scheme: scheme})
  if err != nil {
    return nil, nil, microerror.Mask(err)
  }

  return ctrlClient, clientset, nil
}

// LoginIntoCluster will login into cluster by executing opsctl login command
func loginIntoCLuster(cluster []string) error {
  args := append([]string{"login", "--no-cache"}, cluster...)
  c := exec.Command("opsctl", args...)

  c.Stderr = os.Stderr
  c.Stdin = os.Stdin

  err := c.Run()
  if err != nil {
    return microerror.Mask(err)
  }
  return nil
}

func contextNameFromCluster(cluster []string) string {
  if len(cluster) == 1 {
    return fmt.Sprintf("gs-%s", cluster[0])
  } else {
    return fmt.Sprintf("gs-%s-%s-clientcert", cluster[0], cluster[1])
  }
}

func (c *ManagementCluster) RemoveFinalizerOnNamespace() error {
  var ns v1.Namespace

  ctx := context.TODO()

  err := c.KubernetesClient.Get(ctx, client.ObjectKey{Name: c.Namespace}, &ns)
  if err != nil {
    return microerror.Mask(err)
  }

  finalizers := ns.GetFinalizers()
  if slices.Contains(finalizers, finalizer) {
    index := slices.Index(finalizers, finalizer)

    ns.SetFinalizers(append(finalizers[:index], finalizers[index+1:]...))

    err = c.KubernetesClient.Update(ctx, &ns)
    if err != nil {
      return microerror.Mask(err)
    }
  }

  return nil
}

func (c *ManagementCluster) SetFinalizerOnNamespace() error {
  var ns v1.Namespace

  ctx := context.TODO()

  err := c.KubernetesClient.Get(ctx, client.ObjectKey{Name: c.Namespace}, &ns)
  if err != nil {
    return microerror.Mask(err)
  }

  finalizers := ns.GetFinalizers()
  if ! slices.Contains(finalizers, finalizer) {

    ns.SetFinalizers(append(finalizers, finalizer))

    err = c.KubernetesClient.Update(ctx, &ns)
    if err != nil {
      return microerror.Mask(err)
    }
  }

  return nil
}

func (c *Cluster) FetchApps() ([]apps.App, error) {
  objList := &apps.AppList{}
  ctx := context.TODO()

  // todo: not possible to filter on "spec.catalog" bc/ cached list not indexed?
  selector := client.MatchingFields{"metadata.namespace": c.WcName}
  //selector := client.MatchingLabels{"app.kubernetes.io/name"
  err := c.SrcMC.KubernetesClient.List(ctx, objList, selector)
  if err != nil {
    return nil, microerror.Mask(err)
  }

  return objList.Items, nil
}


