## 8/30 Release

### Features
1. Edge clusters
1. Cascading edge clusters
2. Inter-cluster communication

#### Edge Cluster
The goal of this feature is to allow running K8s-flavored cluster on the edge and can both tolerate network and node failure. In specific:

- K8s cluster running in the cloud as the cloud control plane, while various K8s flavored clusters (e.g. vanilla K8s, Arktos, etc.) running on the edge
  - [ ] Status of edge clusters and edge workloads can be obtained through the cloud control plane via kubectl
  - [ ] Workloads (e.g. deployments, jobs, etc.) run in edge clusters, and workload CRUD can be managed from the cloud
  - [ ] Workloads (Deployment, pod) created from cloud control plane targeting edge clusters are propagated to the edge following filters
- Failure tolerance 
  - [ ] Edge clusters and workloads continue functioning when the network with cloud disconnects
  - [ ] Edge clusters and workloads resume status and workload syncing when network connection recovers
  - [ ] Edge workloads continue functioning if resource allows when edge cluster nodes fail (K8s behavior)

#### Cascading Edge Clusters
The goal of this feature is to allow multiple edge clusters to be cascaded in hierarchical form. In specific:

- [ ] K8s flavored clusters (vanilla K8s, Arktos leaf cluster) running on the edge can be attached to another edge cluster
- [ ] Cluster and workload status are propagated to the cloud control plane in two modes
  - Regular heart beat (longer interval, Mission & edge cluster)
  - Change-based reporting (e.g. every 10s)
- [ ] Edge clusters can be specified by filters representing geological or organizational topology

#### Inter-Cluster Communication
The goal of this feature is to allow workloads (pods) to communicate with each other when distrubted into multiple edge clusters. In specific:
- [ ] A gateway for a edge cluster to establish a communication "channel" with other clusters on the edge 
- [ ] Network workloads (e.g. VPC, subnet) can be created in selected edge clusters
- [ ] (Mizar) Applicatio workloads such as pods can be created into the specified VPC and subnet 
- [ ] (Mizar) Applicatio workloads (pods) in the same edge cluster and VPC could communicate
  - [ ] Both pods in the same subnet
  - [ ] Pods in different subnets
- [ ] Applicatio workloads (pods) in different edge clusters but same the same VPC could communicate
  - [ ] Pods in different subnets
  

