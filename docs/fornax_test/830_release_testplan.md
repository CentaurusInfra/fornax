# Fornax End-to-End Test for 2021-8-30 Release

This test suite verifies the features for 2021-8-30 release. Detailed explanation about these features are given in [design doc](https://github.com/CentaurusInfra/fornax/tree/main/docs/fornax-design/530_design.md). 

The following tests assume four clusters, denoted as A,B,C and D, are created and configured for tests. A,B,C are kubernetes clusters created using kubeadm, while cluster D is an arktos cluster started by running script arktos-up.sh (https://github.com/CentaurusInfra/arktos/blob/master/hack/arktos-up.sh). These clusters are configured in a hierarchical topology, where Cluster B is an edge cluster to Cluster A, C edge to B, and D edge to C. Detailed instructions on how to configure the hierarchical connection among these clusters are given in the doc test_cluster_setup.md in the same directory as this doc. 


## Overview of Test Coverage

The test cases can be classified into three groups. 

**Group I: EdgeCluster Connection**

Test Case 1: Register Hierarchical Edge Clusters with Cloud

Test Case 3: An edge cluster re-connects to the cloud

Test Case 5: An edge cluster continues to work when disconnected from the cloud

Test Case 6: An edge cluster conintues to monintor its sub-edge-clusters when disconnected from cloud



**Group II: Workload Managment**

Test Case 4: Deploy workload (deployment) to edge clusters using mission

Test Case 7: Deploy workload (job) to edge clusters using mission

Test Case 8: Check mission status using command "kubectl get missions" 

Test Case 9: Update workload to edge clusters using mission

Test Case 10: The mission status of a cluster is "cluster unreachable" if cluster disconnected

Test Case 11: Delete workload to edge clusters using mission

Test Case 12: Deploy workload to an edge cluster with a specific name

Test Case 13: Deploy workload to selective edge clusters with given labels

Test Case 15: An edge cluster picks up the new mission added during its disconnection when reconnected to the cloud

Test Case 16: An edge cluster picks up the change when reconnected if the mission is updated during its disconnection

Test Case 17: An edge cluster deletes a mission when reconnected if the mission is deleted during its disconnection 

Test Case 18: The mission content deleted in edge cluster will be re-instated

Test Case 19: The mission content changed in edge cluster will be reverted automatically


**Group III: Montioring EdgeCluster & Mission States**

Test Case 2: Check edge cluster status when disconnected from the cloud 

Test Case 14: Check edge cluster status when connected from the cloud 

Test Case 20: The status of edge cluster is "Unhealthy" if the clusterd is connected but the underlying cluster is unreachable


## Test Cases

**Note: By default, the commands in this doc are run from the root directory of the fornax repo. If not specified, the command is run on the machine of Cluster A, which will be referred as "root operator machine" in the rest of this document.**


**Test Case 1: Register Hierarchical Edge Clusters with Cloud**

Step 1: start kubeedge cloudcore in the root operator machine, using command:
```
_output/local/bin/cloudcore
```

Step 2: Run the following command in the root operator machine to verify that **NO** edge cluster is registered.
```
kubectl get edgeclusters
```

Step 3: in Cluster B machine, start the edgecore in edge-cluster mode, with command 
```
_output/local/bin/edgecore --edgecluster
```


Step 4: wait 10-20 seconds, in the root operator machine, verify that the edge-cluster B shows up in the command ouput of 
```
kubectl get edgeclusters
```

Step 5: in cluster B machine, start cloudcore
```
_output/local/bin/cloudcore
```

Step 6: in Cluster C machine, start the edgecore in edge-cluster mode, with command 
```
_output/local/bin/edgecore --edgecluster
```


Step 7: wait 30-40 seconds, in Cluster A machine, verify that the C is shown as a sub-edge-cluster of edge-cluster B in the command ouput of 
```
kubectl get edgeclusters
```

Step 8: in cluster C machine, start cloudcore
```
_output/local/bin/cloudcore
```

Step 9:in Cluster D machine, start the edgecore in edge-cluster mode, with command 
```
_output/local/bin/edgecore --edgecluster
```


Step 10: wait 50-60 seconds, in Cluster A machine, run command, 
```
kubectl get edgeclusters
```
Verify that the clusters C and C/D (C/D mean D is a sub-cluster of C) are shown as sub-edge-clusters of edge-cluster B in the command ouput of 

**Test Case 2: Check edge cluster status when disconnected from the cloud**

Continuing from the previous test case, do the following:

Step 1: kill the process of edgecore in cluster B machine.

Step 2: wait for 2 minutes. Run the command in the root operator machine,
```
kubectl get edgeclusters
```
Verify the edge cluster B exists but its state is "Disconnected".

Also verify that the info of "SubEdgeClusters" C and C/D is cleared. 

**Test Case 3: An edge cluster re-connects to the cloud**

Continuing from the previous test case, do the following:

Step 1: in cluster B machine, restart the edgecore in edge-cluster mode, using command 
```
edgecore --edgecluster
```

Step 4: wait 10-20 seconds, verify that the state of the edge-cluster B is "healthy" in the command ouput 
```
kubectl get edgeclusters
```

Also verify that the states of cluster C and C/D are displayed in the section of "SubEdgeClusters" in the command output. 

**Test Case 4: Deploy workload (deployment) to edge clusters using mission**

Step 1: Run the command in the root operator machine
```
kubectl apply -f tests/edgecluster/data/missions/deployment-to-all.yaml
```

Step 2: run the following command to verify the mission is deployed and the status of the mission is shown and updated regularly.
```
kubectl get missions
```

Step 3: run the following command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the deployment specified in the mission content is created in the edge clusters of B, C, D. 


**Test Case 5: An edge cluster continues to work when disconnected from the cloud**

Continuing from the previous test case, do the following:

Step 1: kill the process of edgecore in cluster B. 

Step 2: run the following command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the deployment created via missions in the previous test case is still active in each edge cluster. 


**Test Case 6: An edge cluster conintues to monintor its sub-edge-clusters when disconnected from cloud**

Continuing from the previous test case, do the following:

Step 1: Run command 
```
kubectl get edgecluster --kubeconfig=[cluster_B_kubeconfig]
```

Verify the info of edge cluster C is displayed correctly, just like cluster B is not disconnected from the cloud.

Step 2: Run command 
```
kubectl get mission --kubeconfig=[cluster_B_kubeconfig]
```

Verify the info of mission deployment-to-all is displayed correctly, just like cluster B is not disconnected from the cloud.

**Test Case 7: Deploy workload (job) to edge clusters using mission**

Step 1: Run the command in the root operator machine
```
kubectl apply -f tests/edgecluster/data/missions/job-to-all.yaml
```

Step 2: run the following command to verify the mission is deployed
```
kubectl get missions
```

Step 3: run the following command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get jobs --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the job specified in the mission content is created in each edge cluster. 


**Test Case 8: Check mission status using command "kubectl get missions"** 

Continuing from the previous test case,

Step 1: run command and watch the output for 1 minute
```
watch kubectl get missions
```

Step 2: verify the the status of the deployed mission are shown correctly and automatically updated. Make sure the name of the edge cluster (B, B/C, B, C and D) and the status of the mission content(deployment/job) are shown in pairs.

**Test Case 9: Update workload to edge clusters using mission**

Continuing from the previous test case,

Step 1: Note down the number of replicas of the deployment specified in the mission content of tests/edgecluster/data/missions/deployment-to-all.yaml

Step 2: Change the number of deployment replicaset number in the mission content of tests/edgecluster/data/missions/deployment-to-all.yaml.

Step 3: Run and verify the following command returns successfully
```
kubectl apply -f tests/edgecluster/data/missions/deployment-to-all.yaml
```

Step 4:  run the following command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the number of replicas in the deployment created via mission in each edge cluster is updated. 

**Test Case 10: The mission status of a cluster is "cluster unreachable" if cluster disconnected**

Continuing from the previous test case,

Step 1: kill the edgecore process in cluster C and wait for 2 minutes. 

Step 2: run the following command  
```
kubectl get missions
```

verify that:

    a. the mission status of cluster B/C is "cluster unreachable" 

    b. the status of the cluster under B/C, namely B/C/D, is cleared.

**Test Case 11: Delete workload to edge clusters using mission**

Continuing from the previous test case,

Step 1: Run and verify the following command returns successfully
```
kubectl delete -f tests/edgecluster/data/missions/deployment-to-all.yaml
```

Step 2: run the following command  to verify the mission is deleted.
```
kubectl get missions
```

Step 3:  run the following command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the deployment specified in the mission content is gone in each edge cluster. 

**Test Case 12: Deploy workload to an edge cluster with a specific name**

Step 1. update tests/edgecluster/data/missions/deployment-to-given-clusters.yaml to change the value of spec/placement/cluster/Name to be the name of cluster C.

Step 2: Run and verify the following command returns successfully
```
kubectl apply -f tests/edgecluster/data/missions/deployment-to-given-clusters.yaml
```

Step 3: run the following command to verify the mission is deployed in cluster A, B, C and D, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster A, B, C and D respectively, 
```
kubectl get missions --kubeconfig=[edge_cluster_kubeconfig]
```

Step 4: run the following command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively, 
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the deployment specified in the mission content is created in cluster C only. 

Step 5: In the root operator machine, run command:
```
kubectl get mission 
```
Verify that the mission states of cluster B and B/C/D are reported as "not match", while the state of cluster B/C is the status info of the deployment.


**Test Case 13: Deploy workload to selective edge clusters with given labels**

Step 1. Double check the /etc/kubeedge/edgecore.yaml files in the cluster B, C and D and make sure only Cluster D has the label of "company" : "futurewei".

Step 2: Run and verify the following command returns successfully
```
kubectl apply -f tests/edgecluster/data/missions/deployment-to-given-labels.yaml
```

Step 3: run the following command to verify the mission is deployed in cluster A, B, C and D, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster A, B, C and D respectively, 
```
kubectl get missions --kubeconfig=[edge_cluster_kubeconfig]
```

Step 4: run the following command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively, 
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the deployment is created in Cluster D only.

Step 5: In the root operator machine, run command:
```
kubectl get mission 
```
Verify that the mission states of cluster B and B/C are reported as "not match", while the state of cluster B/C/D is the status info of the deployment.


**Test Case 14: The information in the edge cluster status**

Step 1: Run Command in the root operator machine
```
kubectl get edgeclusters
```

Verify the following information of edge cluster B are correctly displayed:

a. LastHeartBeat 

b. HealthStatus: should be "healthy"

c. SubEdgeClusters: should include "C" and "C/D"

d. Received_Missions

e. Matched_Missions

Step 2: Run Command in the root operator machine
```
kubectl get edgeclusters --kubeconfig=[Cluster_B_kubeconfig]
```

Verify the following information of edge cluster C are correctly displayed:

a. LastHeartBeat 

b. HealthStatus: should be "healthy"

c. SubEdgeClusters: should include "D""

d. Received_Missions

e. Matched_Missions

**Test Case 15: An edge cluster picks up the new mission added during its disconnection when reconnected to the cloud**

Continuing from the previous test case, do the following:

Step 1: delete all the missions deployed using command 
```
kubectl delete mission [mission_name]
```

And double check they are all gone using command
```
kubectl get missions
```

Step 2: kill the process of edgecore in cluster B. 

Step 3: Run the following command and verify it returns successfully
```
kubectl apply -f tests/edgecluster/data/missions/deployment-to-all.yaml
```

Step 4: run command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get missions --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the mission is **NOT** created in the edge clusters B, C and D. 


Step 5: run command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the deployment specified in the mission content is **NOT** created in the edge clusters B, C and D. 

Step 6: restart the edgecore in edge-cluster mode and wait 20 seconds
```
edgecore --edgecluster
```

Step 7:  run command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get missions --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the mission is created in the edge clusters B, C and D.  

Step 8:  run command , with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the deployment specified in the mission content is created in the edge clusters B, C and D.  

**Test Case 16: An edge cluster picks up the change when reconnected if the mission is updated during its disconnection**

Continuing from the previous test case, do the following:

Step 1: kill the process of edgecore in Cluster B. 

Step 2: update the number of replicas in deployment spec in the mission content of tests/edgecluster/data/missions/deployment-to-all.yaml.

Step 2: Run the following command and verify it returns successfully
```
kubectl apply -f tests/edgecluster/data/missions/deployment-to-all.yaml
```

Step 4: run command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get missions --kubeconfig=[edge_cluster_kubeconfig] -o json
```
Verify that the mission content is **NOT** updated in the edge clusters B, C and D.  


Step 5: run command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the replica number of the deployment created via mission is **NOT** updated  in the edge clusters B, C and D. 

Step 6: restart the edgecore in edge-cluster mode in cluster B. Wait 20 seconds

Step 7:  run command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get missions --kubeconfig=[edge_cluster_kubeconfig] -o json
```
Verify that the mission content is updated in the edge clusters B, C and D. 

Step 8:  run command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the number of replicas in the deployment is updated in the edge clusters B, C and D. 


**Test Case 17: An edge cluster deletes a mission when reconnected if the mission is deleted during its disconnection **

Continuing from the previous test case, do the following:

Step 1: kill the process of edgecore in cluster B. 
 
Step 2: Run the following command and verify it returns successfully
```
kubectl delete -f tests/edgecluster/data/missions/deployment-to-all.yaml
```

Step 4: run command , with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get missions --kubeconfig=[edge_cluster_kubeconfig] -o json
```
Verify that the mission content is **NOT** deleted in the edge clusters B, C and D.  


Step 5: run command, with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the replica number of the deployment created via mission is still active in the edge clusters B, C and D. 

Step 6: restart the edgecore in edge-cluster mode in cluster B and wait 20 seconds

Step 7:  run command with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get missions --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the mission is deleted in the edge clusters B, C and D. 

Step 8:  run command with [edge_cluster_kubeconfig] set to the kubeconfig files of cluster B, C and D respectively,
```
kubectl get deployment --kubeconfig=[edge_cluster_kubeconfig]
```
Verify that the the deployment created via mission is gone in the edge clusters B, C and D.  


**Test Case 18: The mission content deleted in edge cluster will be re-instated**

Step 1: Run the following command to deploy the mission
```
kubectl apply -f tests/edgecluster/data/missions/deployment-to-all.yaml
```


Step 2: run command verify the deployment specified in the mission is deployed in the edge cluster B
```
kubectl get deployment --kubeconfig=[cluster_B_kubeconfig]
```

Step 3:  run command to delete the deployment
```
kubectl delete deployment [deployment_name] --kubeconfig=[cluster_B_kubeconfig]
```


Step 4: verify the deployment specified in the mission is deleted (Note: do it immediately after the previous step)
```
kubectl get deployment --kubeconfig=[cluster_B_kubeconfig]
```

Step 8:  Wait 20 seconds, run command to verify the deployment specified in the mission is back
```
kubectl get deployment --kubeconfig=[cluster_B_kubeconfig]
```


**Test Case 19: The mission content changed in edge cluster will be reverted automatically**

Step 1: Run the following command to deploy the mission
```
kubectl apply -f tests/edgecluster/data/missions/deployment-to-all.yaml
```


Step 2: run command and verify the deployment specified in the mission is deployed in the edge cluster B
```
kubectl get deployment --kubeconfig=[cluster_B_kubeconfig]
```

Step 3:  run command to scale the deployment
```
kubectl scale deployment.v1.apps/[deployment_name] -rreplicas=10 --kubeconfig=[cluster_B_kubeconfig]
```


Step 4: verify the number of replicas in deployment specified in the mission is changed to 10 (Note: do it immediately after the previous step)
```
kubectl get deployment --kubeconfig=[cluster_B_kubeconfig]
```

Step 8:  Wait 20 seconds, verify the number of the replicas in the deployment specified in the mission is reverted to the original value
```
kubectl get deployment --kubeconfig=[cluster_B_kubeconfig]
```

**Test Case 20: The status of edge cluster is "Unhealthy" if the clusterd is connected but the underlying cluster is unreachable**

Step 1: Stop the arktos-up.sh script in cluster D.

Step 2: make sure the edgecore in cluster D is still running

step 3: Wait 20 seconds and check the "HealthStatus" of Cluster D in the output of the following command is "Unhealthy". 
```
kubectl get edgeclusters --kubeconfig=[cluster_C_kubeconfig]
```

step 4: Run the command in the root operator machine. 
```
kubectl get edgeclusters 
```
Verify the in the infor of EdgeCluster B, the state of SubEdgeCluster C/D is "unhealthy".