package cluster

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fatih/color"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *Cluster) ApplyCAPIApps(filename string) error {
	// we skip the app apply if the file is empty
	fileInfo, err := os.Stat(c.AppYamlFile(filename))
	if err != nil {
		return microerror.Mask(err)
	}
	// Check if the file size is 0
	if fileInfo.Size() == 0 {
		return microerror.Maskf(MigrationFileEmpty, "Migration File is empty. Nothing to migrate")
	}

	// waitloop til kubeconfig/default-cluster-values are found
	for {
		cmClusterValuesExists, err := checkIfObjectExists(c.DstMC.KubernetesClient, c.OrgNamespace, fmt.Sprintf("%s-cluster-values", c.WcName), "configmap")
		if err != nil {
			fmt.Printf("Error checking existence of %s/%s-cluster-values: %s\n", "configmap", c.WcName, err)
			time.Sleep(time.Second * 5)
			continue
		}

		if cmClusterValuesExists {
			secretClusterValuesExists, err := checkIfObjectExists(c.DstMC.KubernetesClient, c.OrgNamespace, fmt.Sprintf("%s-cluster-values", c.WcName), "secret")
			if err != nil {
				fmt.Printf("Error checking existence of %s/%s-cluster-values: %s\n", "secret", c.WcName, err)
				time.Sleep(time.Second * 5)
				continue
			}

			if secretClusterValuesExists {
				kubeconfigExists, err := checkIfObjectExists(c.DstMC.KubernetesClient, c.OrgNamespace, fmt.Sprintf("%s-kubeconfig", c.WcName), "secret")
				if err != nil {
					fmt.Printf("Error checking existence of %s/%s-kubeconfig: %s\n", "secret", c.WcName, err)
					time.Sleep(time.Second * 5)
					continue
				}

				if kubeconfigExists {
					color.Yellow("\nAll prerequistes are found on the new MC for app migration")
					break
				}
			}
		}
	}

	fmt.Printf("Applying all non-default APP CRs to MC\n")
	applyManifests := func() error {
		//nolint:gosec
		e := exec.Command("kubectl", "--context", fmt.Sprintf("gs-%s", c.DstMC.Name), "apply", "-f", c.AppYamlFile(filename))

		e.Stderr = os.Stderr
		e.Stdin = os.Stdin

		err := e.Run()
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	}

	err = backoff.Retry(applyManifests, c.BackOff)
	if err != nil {
		return microerror.Mask(err)
	}
	color.Green("All non-default apps applied successfully.\n\n")
	return nil
}

func checkIfObjectExists(k8s client.Client, nameSpace string, name string, resourceKind string) (bool, error) {
	switch resourceKind {
	case "secret":
		var secret v1.Secret
		err := k8s.Get(context.TODO(), client.ObjectKey{
			Name:      name,
			Namespace: nameSpace,
		},
			&secret)

		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil

	case "configmap":
		var cm v1.ConfigMap
		err := k8s.Get(context.TODO(), client.ObjectKey{
			Name:      name,
			Namespace: nameSpace,
		},
			&cm)

		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil

	default:
		return false, fmt.Errorf("unsupported resource kind: %s", resourceKind)
	}
}
