# Test Cluster Setup For 2021-8-30 Release Fornax End-to-End Test

This doc describeds how to setup test clusters for the 2021-8-30 Fornax release test, whose test cases are given in the release_testplan.md in the same folder as this doc. 

This test requires four clusters, denoted as A,B,C and D. A,B,C are kubernetes clusters created using kubeadm, while cluster D is an arktos cluster started by running script arktos-up.sh (https://github.com/CentaurusInfra/arktos/blob/master/hack/arktos-up.sh). These clusters are configured in a hierarchical topology, where Cluster B is an edge cluster to Cluster A, C edge to B, and D edge to C. 

Machine A is referred as the "root operator machine" in these two docs.

**Note: Run the commands in these two docs as a root user**

## Machine Preparation

1. Prepare 4 AWS machines, t2.xlarge, 80G storage, ubuntu 18.04, for the clusters of A, B, C and D.

2. Open the port of 10000 & 10002 in the security group of of machine A, B and C.

3. Open the port of 6443 in the security group of of machine A, B, C and D.

4. In machine A, B, C, create a Kubernetes cluster following doc https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/.

5. In machine D, clone the repo https://github.com/CentaurusInfra/arktos/ and start an Arktos cluster byt running the script arktos-up.sh.

## Kubeedge Configuration

### Kubeonfig File Preparation

Copy the admin kubeconfig file of cluster A to machine B, the kubecofig file of cluster B to the machine of cluster C, and the kubeconfig file of cluster C to the machine of cluster D.

Copy the kubeconfig files of cluster A, B, C and D to the root operator machine.

### In machine A, do the following


1. Clone a repo of https://github.com/CentaurusInfra/fornax, sync to the branch/commit to test.

Build the binaries of edgecore and cloudcore using the commands
```
make WHAT=cloudcore
make WHAT=edgecore
```

2. config cloudcore 
```
cp /etc/kubernetes/admin.conf /root/.kube/config
_output/local/bin/cloudcore --minconfig > /etc/kubeedge/config/cloudcore.yaml
```

3. Generate security data 

Note down the IP address of machine A, B, C, and D, denotes as IP_A, IP_B, IP_C and IP_D, and run the command:

```
build/tools/certgen.sh genCA IP_A IP_B IP_C IP_D
build/tools/certgen.sh genCertAndKey server IP_A IP_B IP_C IP_D
```

Then copy the files of folder /etc/kubeedge/ca and /etc/kubeedge/certs in machine A to the folder of /etc/kubeedge/ca and /etc/kubeedge/certs in machine B, C and D. 

5. Install CRDs
```
export KUBECONFIG=[Cluster_A_kubeconfig_file]

kubectl apply -f build/crds/devices/devices_v1alpha2_device.yaml
kubectl apply -f build/crds/devices/devices_v1alpha2_devicemodel.yaml 

kubectl apply -f build/crds/reliablesyncs/cluster_objectsync_v1alpha1.yaml
kubectl apply -f build/crds/reliablesyncs/objectsync_v1alpha1.yaml 

kubectl apply -f  build/crds/router/router_v1_rule.yaml
kubectl apply -f  build/crds/router/router_v1_ruleEndpoint.yaml

kubectl apply -f build/crds/edgecluster/mission_v1.yaml
kubectl apply -f build/crds/edgecluster/edgecluster_v1.yaml

```


### In machine B
1. Clone a repo of https://github.com/CentaurusInfra/fornax, sync to the branch/commit to test.

Build the binaries of edgecore and cloudcore using the commands
```
make WHAT=cloudcore
make WHAT=edgecore
```

2. config cloudcore 
```
cp /etc/kubernetes/admin.conf /root/.kube/config
_output/local/bin/cloudcore --minconfig > /etc/kubeedge/config/cloudcore.yaml
```
3. config edgecore
```
cp [Cluster_B_kubeconfig_file] /root/edgecluster.kubeconfig
_output/local/bin/edgecore --edgeclusterconfig > /etc/kubeedge/config/edgecore.yaml
tests/edgecluster/hack/update_edgecore_config.sh [cluster_A_kubeconfig_file]
```

### In machine C
1. Clone a repo of https://github.com/CentaurusInfra/fornax, sync to the branch/commit to test.

Build the binaries of edgecore and cloudcore using the commands
```
make WHAT=cloudcore
make WHAT=edgecore
```

2. config cloudcore 
```
cp /etc/kubernetes/admin.conf /root/.kube/config
_output/local/bin/cloudcore --minconfig > /etc/kubeedge/config/cloudcore.yaml
```
3. config edgecore
```
cp [Cluster_C_kubeconfig_file] /root/edgecluster.kubeconfig
_output/local/bin/edgecore --edgeclusterconfig > /etc/kubeedge/config/edgecore.yaml
tests/edgecluster/hack/update_edgecore_config.sh [cluster_B_kubeconfig_file]
```

### In machine D
1. Clone a repo of https://github.com/CentaurusInfra/fornax, sync to the branch/commit to test.

Build the binary of edgecore 
```
make WHAT=edgecore
```

2. config edgecore
```
cp [Cluster_D_kubeconfig_file] /root/edgecluster.kubeconfig
_output/local/bin/edgecore --edgeclusterconfig > /etc/kubeedge/config/edgecore.yaml
tests/edgecluster/hack/update_edgecore_config.sh [cluster_C_kubeconfig_file]
```

update the /etc/kubeedge/config/edgecore.yaml so the section of spec/clusterd looks like the following:

```
  clusterd:
    ...
    kubeDistro: arktos
    labels:
      "company" : "futurewei"
```
