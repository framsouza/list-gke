# List of Kubernetes Clusters running on GKE

This tool will return the List of GKE cluster runnings on GKE, Node Count and Machine Type

### Usage
First, make sure you have the `GOOGLE_APPLICATION_CREDENTIALS` environment variable set as per the google docs, https://cloud.google.com/docs/authentication/production

It will require two arguments, `project` and `zone`

```
sudo ./list-gke -project=<PROJECTNAME> -zone=<ZONE>
```

Example output:

```
CLUSTER NAME    NODE COUNT      MACHINE TYPE
fred-k8s        3               e2-standard-4
gke             4               e2-standard-4  
```
