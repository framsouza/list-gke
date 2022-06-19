# List of Kubernetes Clusters running on GKE

This tool will return the List of GKE cluster runnings on GKE, Node Count and Pod count.

### Usage
First, make sure you have the `GOOGLE_APPLICATION_CREDENTIALS` environment variable set as per the google docs, https://cloud.google.com/docs/authentication/production

It will require two arguments, `project` and `zone`. By default the zone is set to `-`, which means all zones will be retrieved. 

```
sudo ./list-gke -project=<PROJECTNAME> 
```

Example output:

```
CLUSTER NAME    NODE COUNT      POD COUNT
gke             7               42
fram-k8s        3               23
my-gke          3               22
testing-k8s     3               28
```
