## 8/30 Release

### Features
1. Edge clusters
1. Cascading edge clusters
2. Inter-cluster communication

#### Edge Cluster
The goal of this feature is to allow running K8s-flavored cluster on the edge and can both tolerate network and node failure. In specific:

- Arktos cluster running in the cloud as the cloud control plane
- Various K8s flavored clusters (vanilla K8s, Arktos & K3s) running on the edge (local sites)
- Status of edge clusters and edge workloads can be obtained through the cloud control plane (kubectl, web console, etc.)
- Workloads run in edge clusters, and workload CRUD can be managed from the cloud
- Workloads (Deployment, pod) created from cloud control plane targeting edge clusters are propagated to the edge following filters
- Robustness
  - Edge clusters and workloads continue functioning when the network with cloud disconnects
  - Edge clusters and workloads resume status and workload syncing when network connection recovers
  - Edge workloads continue functioning if resource allows when edge cluster nodes fail (K8s cluster behavior)

#### Cascading Edge Cluster
The goal of this feature is to allow multiple edge clusters to be cascaded in hierarchical form. In specific:

- K8s flavored clusters (vanilla K8s, Arktos & K3s) running on the edge can be attached to another edge cluster
- Edge clusters can be specified by filters representing geological or organizational topology
- Cluster and workload status are propagated to the cloud control plane in two modes
  - Regular heart beat
  - Change-based reporting

#### Edge Cluster Communication
The goal of this feature is to allow workloads (pods) to communicate with each other when distrubted into multiple edge clusters. In specific:
- Pods from the same VPC and in different physical clusters to communicate using Pod IPs

