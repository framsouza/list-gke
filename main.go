package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"text/tabwriter"

	"golang.org/x/oauth2/google"
	container "google.golang.org/api/container/v1"
)

var (
	projectID = flag.String("project", "", "Enter the Project ID")
	zone      = flag.String("zone", "", "Enter the Compute zone")
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

	list, err := svc.Projects.Zones.Clusters.List(projectID, zone).Do()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %v", err)
	}

	fmt.Fprintf(w, "%s\t%s\t%s\t", "CLUSTER NAME", "NODE COUNT", "MACHINE TYPE")

	for _, v := range list.Clusters {
		fmt.Fprintf(w, "\n%s\t", v.Name)
		fmt.Fprintf(w, "%d\t", v.CurrentNodeCount)

		poolList, err := svc.Projects.Zones.Clusters.NodePools.List(projectID, zone, v.Name).Do()
		if err != nil {
			return fmt.Errorf("failed to list node pools for cluster %q: %v", v.Name, err)
		}
		for _, np := range poolList.NodePools {
			fmt.Fprintf(w, "%s\t\t", np.Config.MachineType)
		}

	}

	return nil
}
