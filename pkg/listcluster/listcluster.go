package listClusters

import (
	"context"
	"fmt"
	"os"
	"sort"

	"text/tabwriter"

	container "google.golang.org/api/container/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"

	"github.com/framsouza/list-gke/pkg/kubectl"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

func ListClusters(svc *container.Service, projectID, zone string) error {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	// retriving cluster list
	list, err := svc.Projects.Zones.Clusters.List(projectID, zone).Do()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %v", err)
	}

	// header
	fmt.Fprintf(w, "%s\t\t%s\t\t%s\t", "CLUSTER NAME", "NODE COUNT", "POD COUNT")

	// sorting by node count
	sort.Slice(list.Clusters, func(i, j int) bool {
		return list.Clusters[i].CurrentNodeCount > list.Clusters[j].CurrentNodeCount
	})

	kubeConfig, err := kubectl.GetK8sClusterConfigs(context.TODO(), projectID)
	if err != nil {
		return err
	}

	// gathering cluster name, node acount, number of pods running and machine type
	for _, v := range list.Clusters {
		fmt.Fprintf(w, "\n%s\t\t", v.Name)
		fmt.Fprintf(w, "%d\t\t", v.CurrentNodeCount)

		cfg, err := clientcmd.NewNonInteractiveClientConfig(*kubeConfig, v.Name, &clientcmd.ConfigOverrides{CurrentContext: v.Name}, nil).ClientConfig()
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes configuration cluster=%s: %w", v.Name, err)
		}

		k8s, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client cluster=%s: %w", v.Name, err)
		}

		pods, _ := k8s.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		fmt.Fprintf(w, "%d\t", len(pods.Items))

	}

	return nil
}
