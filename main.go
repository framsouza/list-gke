package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"text/tabwriter"

	"golang.org/x/oauth2/google"
	container "google.golang.org/api/container/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	projectID = flag.String("project", "", "Enter the Project ID")
	zone      = flag.String("zone", "-", "Enter the Compute zone")
)

func main() {
	flag.Parse()

	if *projectID == "" {
		fmt.Fprintln(os.Stderr, "Missing project")
		flag.Usage()
		os.Exit(2)
	}
	if *zone == "" {
		fmt.Fprintln(os.Stderr, "Missing zone")
		flag.Usage()
		os.Exit(2)
	}

	ctx := context.Background()

	hc, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	svc, err := container.New(hc)
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	if err := listClusters(svc, *projectID, *zone); err != nil {
		log.Fatal(err)
	}
}

func listClusters(svc *container.Service, projectID, zone string) error {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	// retriving cluster list
	list, err := svc.Projects.Zones.Clusters.List(projectID, zone).Do()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %v", err)
	}

	// header
	fmt.Fprintf(w, "%s\t%s\t%s\t", "CLUSTER NAME", "NODE COUNT", "POD COUNT")

	// sorting by node count
	sort.Slice(list.Clusters, func(i, j int) bool {
		return list.Clusters[i].CurrentNodeCount > list.Clusters[j].CurrentNodeCount
	})

	kubeConfig, err := getK8sClusterConfigs(context.TODO(), projectID)
	if err != nil {
		return err
	}

	// gathering cluster name, node acount, number of pods running and machine type
	for _, v := range list.Clusters {
		fmt.Fprintf(w, "\n%s\t", v.Name)
		fmt.Fprintf(w, "%d\t", v.CurrentNodeCount)

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

		//machinetype, _ := k8s.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: "beta.kubernetes.io/instance-type"})

		//		"beta.kubernetes.io/instance-type

		//		poolList, err := svc.Projects.Zones.Clusters.NodePools.List(projectID, v.Zone, v.Name).Do()
		//		if err != nil {
		//			return fmt.Errorf("failed to list node pools for cluster %q: %v", v.Name, err)
		//		}
		//		for _, np := range poolList.NodePools {
		//			fmt.Fprintf(w, "%s\t\t", np.Config.MachineType)
		//		}

	}

	return nil
}

// access kube config
func getK8sClusterConfigs(ctx context.Context, projectID string) (*api.Config, error) {
	svc, err := container.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("container.NewService: %w", err)
	}

	// Basic config structure
	ret := api.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters:   map[string]*api.Cluster{},  // Clusters is a map of referencable names to cluster configs
		AuthInfos:  map[string]*api.AuthInfo{}, // AuthInfos is a map of referencable names to user configs
		Contexts:   map[string]*api.Context{},  // Contexts is a map of referencable names to context configs
	}

	// Ask Google for a list of all kube clusters in the given project.
	resp, err := svc.Projects.Zones.Clusters.List(projectID, "-").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("clusters list project=%s: %w", projectID, err)
	}

	for _, f := range resp.Clusters {
		name := fmt.Sprintf(f.Name)
		cert, err := base64.StdEncoding.DecodeString(f.MasterAuth.ClusterCaCertificate)
		if err != nil {
			return nil, fmt.Errorf("invalid certificate cluster=%s cert=%s: %w", name, f.MasterAuth.ClusterCaCertificate, err)
		}

		ret.Clusters[name] = &api.Cluster{
			CertificateAuthorityData: cert,
			Server:                   "https://" + f.Endpoint,
		}

		ret.Contexts[name] = &api.Context{
			Cluster:  name,
			AuthInfo: name,
		}
		ret.AuthInfos[name] = &api.AuthInfo{
			AuthProvider: &api.AuthProviderConfig{
				Name: "gcp",
				Config: map[string]string{
					"scopes": "https://www.googleapis.com/auth/cloud-platform",
				},
			},
		}
	}

	return &ret, nil
}
