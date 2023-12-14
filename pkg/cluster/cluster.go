package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/giantswarm/microerror"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//giantswarmawsalpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
 // clientgoscheme "k8s.io/client-go/kubernetes/scheme"

  	chart "github.com/giantswarm/apiextensions-application/api/v1alpha1"


)

var (
	scheme = runtime.NewScheme()
)

func init() {
//	_ = clientgoscheme.AddToScheme(scheme)
//	_ = capi.AddToScheme(scheme)
//	_ = giantswarmawsalpha3.AddToScheme(scheme)
//	_ = kubeadmv1beta1.AddToScheme(scheme)
	_ = chart.AddToScheme(scheme)
}

type Cluster struct {
	WcName string

	SrcMC *ManagementCluster
	DstMC *ManagementCluster
}

type ManagementCluster struct {
  Name      string
	Namespace string

	KubernetesClient client.Client
}

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
