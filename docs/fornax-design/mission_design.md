# Design of Mission in Fornax

A CRD called "Mission" is used to store the actual workload definition (Mission/MissionSpec/MssionResource or Mission/MissionSpec/MissionCommand) and destination information (Mission/MissionSpec/Placement) to deploy to the edge clusters. User can also configure how the mission state is reported by specifying the field of Mission/StateCheck/Command.

```golang
// Mission specifies a workload to deploy in edge clusters
type Mission struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec MissionSpec `json:"spec"`
	State map[string]string `json:"state,omitempty"`
}

type MissionSpec struct {
	MissionResource string `json:"missionresource,omitempty"`
	MissionCommand MissionCommandSpec `json:"missioncommand,omitempty"`
	Placement GenericPlacementFields `json:"placement,omitempty"`
	StateCheck StateCheckFields `json:"statecheck"`
}

type MissionCommandSpec struct {
	Trigger string `json:"trigger,omitempty"`
	RunWhenTriggerSucceed bool `json:"runwhentriggersucceed"`
	Command string `json:"command,omitempty"`
	ReverseCommand string `json:"reversecommand,omitempty"`
}

type GenericClusterReference struct {
	Name string `json:"name"`
}

type GenericPlacementFields struct {
	Clusters    []GenericClusterReference `json:"clusters,omitempty"`
	MatchLabels map[string]string         `json:"matchlabels,omitempty"`
}
type StateCheckFields struct {
	Command string `json:"command,omitempty"`
}
```

## Mission/MissionSpec/MssionResource
This field is the content of a Kubernetes resource in a yaml file. 

Fornax will
1. use command "kubectl apply -f" to deploy the mission resource, 
2. use command "kubectl delete -f" to delete the mission resource, 
3. use command "kubectl get [resource_type] [resource_name]" to grab the state of mission content, if the mission does not specify the mission state check command. 

Note that Fornax prefers MissionResource over MissionCommand. Users are encoraged to use MissionResource to describe the mission content as long as it is possible, and use the MissionCommand only if they have to. We get the most of the nice feature of declarative object management in Kubernetes if we delcare the mission content using MissionResource.

If both MissionResource and MissionCommand fields are specified in a mission, the MissionCommand will be ignored. A mission is invalid if neither issionResource nor MissionCommand field is specified.

## Mission/MissionSpec/MssionCommand

We use MissionCommand if MissionResource is insufficient, for example:
1. We need to run a kubectl command other than "kkubectl apply" to do the work.
2. Or, we need to do something via a bash command

The Content of the MssionCommand is defined via multiple fields:
1. Trigger: the command to check whether or not we need to run the mission command. By default, the mission command will be run only if the trigger command fails.
2. RunWhenTriggerSucceed: by default it is false, which means the mission command will be run only if the trigger command fails. The mission command will run if the trigger command succeeds, if this value is set to true.
3. Command: the command to deploy the mission content.
4. ReverseCOmmand: the command to delete the mission content. 

Fornax periodically checks the trigger condition and re-run the mission command if necessary. The trigger check is skipped and the mission command will be run if the Trigger command is not defined. However, users are strongly suggested against such a pratice. Mission command usually makes changes to the system and blindly re-run such commands periodically is not a desirable behavior.

### background mission commmand
Fornax assumes that the mission command can be completed within 10 seconds ( this value is not configurable for now. We will change it in the future. ). The command fails if it does not return after 10 seconds.

However, there are some cases where the commmand need to run for good. For example, you might need to keep the "kubectl port-forward" command in a running state as long as the app is running. For such cases, append "&" to the command, which is the same practice to run a background job in Linux. 

## Mission/MissionSpec/Placement
This field specifies which clusters should deploy the mision content. A mission should be deployed in any edge clustersif this field is empty.

There are two ways to restrict the mission content deployment to given edge clusters:
1. if Placement/Clusters field is defined, the mission content will only be deployed to clusters with the names specified.
1. if Placement/MatchLabels field is defined, the mission content will only be deployed to clusters with at least one matching label. Work will be done to use the Kubernetes LabelSelector to defin the matching conditions, which is tracked by https://github.com/CentaurusInfra/fornax/issues/37.

Users can define the labels of the edgecluster in the edgecore config file, which is /etc/kubeconfig/edgecore.yaml by default. 

The following built-in labels are defined by Fornax automatically:
1. kubernetes.io/hostname, the value is the master node host name.
2. role.kubernetes.io/edgecluster, the value is empty
3. edgeclusters.kubeedge.io/kubedistro, the value is the Kubernetes distro of the edge cluster, such as k8s, arktos, etc. 


## Mission Examples
Some simple mission yamls are given in tests/edgecluster/data/missions.

The mission yamls to build a more complicated AI face recognition app is given in tests/edgecluster/data/ai_app/.
