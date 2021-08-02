# Centaurus - Edge + SIG Weekly Meeting
- Wednesday 3:00 - 4:00 PM
- [ZOOM](https://futurewei.zoom.us/j/93051877352?from=addon)


## 02/17

### Agenda:

- Tracks
  - Research
  - Code (Extend Arktos to deploy to Edge https://github.com/pdgetrf/ArktosEdge/issues/3)
- Edge project tracking repo
  - [Project Tracking](https://github.com/pdgetrf/ArktosEdge/projects/1)
- KubeEdge discussion
  - Setup & play ([same network](https://github.com/pdgetrf/ArktosEdge/wiki/KubeEdge-Setup-(Same-Network)), [cloud in GCP, edge in AWS](https://github.com/pdgetrf/ArktosEdge/wiki/KubeEdge-Setup-(Different-Network)))
  - Code walkthrough & deep(er) dive
- Research
  - [AWS Innovate](https://aws.amazon.com/events/aws-innovate/machine-learning/?sc_channel=em&sc_campaign=APAC_FIELD_T1_en-innovate-aiml_20210224_7014z000001MMP3&sc_publisher=aws&sc_medium=em_aws_innovate_aiml&sc_content=field_t1event_field&sc_country=mult&sc_geo=mult&sc_category=mult&sc_outcome=field&trkCampaign=innovate-ml&trk=em_regconfirmed_innovateml21)
  - [State of AI 2020 Report](https://www.stateof.ai), also a [discussion of it](https://youtu.be/o2fYsrV-YlQ)
  - Topics:
    - Scheduling, PolarisCloud
    - Model deployment ([A good demo using GCP](https://youtu.be/fw6NMQrYc6w))
    - Scalability
    - K3s vs KubeEdge

### Notes:
- [x] Global scheduling vs Edge scheduling & placement
  - Scheduling algorithm
- [x] Collect list of ideas for discussion
- [x] Schedule a meet about PolarisCloud 
- [x] Edge models : KubeEdge vs Light mode KubeEdge vs K3s
- [ ] Security vs latency, websocket or better options.
- [ ] Scale up from 2,3k 
- [x] KubeFlow & https://github.com/kubeflow/kfserving

## 02/24

### Agenda
- Edge solution survey (tracking by https://github.com/pdgetrf/ArktosEdge/issues/9)
  - Single node edge
  - Autonomous cluster at Edge (see slides)
  - Features
    - cloud <-> edge tunnel/proxy
    - edge anomomy
      - edge-side apiserver/caching layer
      - edge node and npod health check / stay alive
      - resource grouping (node group, cluster group, deploymet/service group)
    - edge cluster
      - deployment and management
      - self-managing cluster, master migration on power loss

- Other tracking topics
  - KubeEdge MEC Architectur (Mizar into EdgeMesh, gateway, Edge-edge comm, Deepak 3/3 talk)
  - RAINBOW Fog-aware Scheduler


- New Finds
  - [CNCF Cloud Native Interactive Landscape](https://landscape.cncf.io/)
  - [LF AI & Data Foundation Interactive Landscape](https://landscape.lfai.foundation/)
  - [Looks familiar?](https://www.google.com/url?sa=i&url=https%3A%2F%2Fwww.alamy.com%2Fstock-photo-rear-of-a-subaru-imprezza-wrx-wagon-covered-in-bumperstickers-with-19518801.html&psig=AOvVaw3uMN5aEUxpywIhzk69H26W&ust=1614118138974000&source=images&cd=vfe&ved=0CAIQjRxqFwoTCJiRl-XA_u4CFQAAAAAdAAAAABAO)


### Notes
- [ ] lite edge vs full fledged

## 03/03

### Agenda
- EdgeCluster feature
  - Edge capability: survive both network and node failure
  - Locality is important
  - Self-managing cluster
  - EdgeCluster as a custom resource

## 03/10

### Agenda
- n-ary tree based Arktos Edge topology

### Notes
- [x] complexity worth it?
- [ ] local firewall preventing clusters to connect to their root clusers
- [x] local node vs gloal management
- [x] multi-tenancy at edge node (in recursive case, there's a conflict case to discuss)
- [x] use case scenario for hireachy structure
  - [x] kube MEC meeting
- [x] centralized controls over de-centralized control
- [x] schedule meeting with Deepak about network and security


## 03/18

### Agenda
- hireachy model & use cases
- cluster connector demo
- multi-tenancy EdgeCluster
- EdgeCluster provisioning model
- EdgeCluster components initial thoughts (cut short due to time limit)

### Notes
- EdgeCluster as first release goal
- VPC in EdgeCluster needs further looked into

## 03/24

### Agenda
- Scoping: 
  - [Virtual kubelet](https://virtual-kubelet.io/docs/architecture/)
  - [Azure IoT deployment](https://azure.microsoft.com/en-us/blog/manage-azure-iot-edge-deployments-with-kubernetes/)
  - [ioFog](https://iofog.org/docs/2/getting-started/core-concepts.html)
  - Federation & [cluster registry](https://github.com/kubernetes/cluster-registry)
  - Networking NFV, SDN, K8s networking
  - Cluster cascading
- Milestone draft
- Design focus: EdgeCluster upstream and downstream
- Demo by Qian

### Notes
argument between all resource connecting to cloud vs local inter-connected edge cluster

## 03/31
### Agenda
- 4/27 interval SIG review
- New KubeCon NA CFP deadline: [5/23](https://events.linuxfoundation.org/kubecon-cloudnativecon-north-america/program/cfp/)
- Feature Scoping summary
- Settle the debate from last meeting (by showing them under a single framework)
- AWS Outputs/Local Zones/Wavelength (and assumption of how they work)

### Notes
- SOW plan (in a week or two)
- Release in 5/30 or 6/30

## 4/7
### Agenda
- POC design and progress
  - KubeEdge
  - OpenYurt 
- KubeEdge MEC planning

### Notes

# 4/14
### Agenda
- Community meeting  
- POC design and progress
  - KubeEdge
  - OpenYurt (in details by Qian)
- KubeEdge MEC topics

### Notes
- POC 
- Vision doc


# 4/21
### Agenda
- Design Communication
  - Team demo by Qian
  - Discussion with Liguang
  - TSC meeting on 27th (what to bring?)
- POC 
  - Dev env setup
  - Merge KubeEdge with Arktos
  - Add edge cluster support
    - Top down route: Add CRDs to cluster
    - Bottown up route: Extend edged to accept and forward mission (workload) request to Arktos
- KubeEdge community meeting study
   - Edge cluster
   - AKE (support kubectl on edge node)
- Inter-edgecluster communication
- VR MEC demo by AWS Wavelength
- KubeEdge MEC topics

### Notes
1. inter-EdgeCluster 
2. SOW draft
3. Schedule meeting to sync with KubeEdge meeting
4. 07/31 first release, 05/31 design deadline

# 4/28
### Agenda
- Overall progress review
- POC progress update
- Global Scheduling project handoff

# 5/5
### Agenda
- Overall progress review
- Global Scheduling project handoff

- POC progress update
  - cascading cluster demo by Qian
    - client-go vs kubectl
  - Status reporting POC plan
- SOW ideas
  - client-go
  - cluster/workload status update
- Global scheduling for the Edge
- MEC status
- Meeting consolidation
  - Monday night: Global Scheduler
  - Monday & Wednesday: Edge scrum
  - Tuesday night: Arktos community
  - Wednesday afternoon: Edge SIG
  - Thursday night: KubeEdge, KubeEdge MEC

# 5/12
### Agenda
- Overall progress review
- Global Scheduler project
- POC focus
  - Edge cluster status
  - inter Edge cluster comm
  - Self-organizing cluster
  - Scheduler enhancement
- Edge cluster status update design
- MEC

### Note
- Deployment vs Mission
- Status needs to be stored on top layers
- First release scope (2 layers)
- Cluster status in 2nd release
- Workload status in 1st release. Periodic update first. 

# 5/19
### Agenda
- Overall progress review
- POC
  - [x] Cascading cluster
  - [x] Edge cluster status
  - [x] inter Edge cluster comm
  - [ ] Edge core refactor into edge agent (working with KubeEdge)
  - [ ] Self-organizing cluster
  - [ ] Scheduler enhancement
- Demo & Perf
  - Edge application to showcase POC
- Publication
  - KubeCon (due by 5/23)
  - Open Networking & Edge Summit + Kubernetes on Edge Day ([Due by 6/20](https://events.linuxfoundation.org/open-networking-edge-summit-north-america/program/cfp/))
  - Community White paper
- MEC

### Note


# 5/26
### Agenda
- Overall progress review
- POC
  - [x] Cascading cluster
  - [x] Edge cluster status
  - [ ] inter Edge cluster comm
  - [ ] Edge core refactor into edge agent (working with KubeEdge)
  - [ ] Self-organizing cluster
  - [ ] Scheduler enhancement
- Demo & Perf
  - [ ] Edge application to showcase POC
- Publication
  - [x] KubeCon (due by 5/23)
  - [ ] Open Networking & Edge Summit + Kubernetes on Edge Day ([Due by 6/20](https://events.linuxfoundation.org/open-networking-edge-summit-north-america/program/cfp/))

### Note


# 6/2
### Agenda
- Edge Status
  - [x] Cascading cluster
  - [x] Edge cluster status
  - [ ] inter Edge cluster comm (Qian, Shaojun)
    - Mizar setup
    - Zeta
    - [libp2p](https://libp2p.io)
    - Application load balancer
  - [ ] Self-organizing cluster (Eunju)
  - [ ] Env setup doc, repository, CICD, etc.
  - [ ] Slack edge-dev channel
  - [ ] Edge core refactor into edge agent (working with KubeEdge)
  - [ ] Scheduler enhancement
- Publication
  - [x] KubeCon (due by 5/23)
  - [ ] Open Networking & Edge Summit + Kubernetes on Edge Day ([Due by 6/20](https://events.linuxfoundation.org/open-networking-edge-summit-north-america/program/cfp/))
  - [ ] Design doc (due by 6/4, Peng & Qian) 
- Demo & Perf
  - [ ] Edge application to showcase POC
- MEC
  - Dev & Release plan 
    - Akrano demo 8/26


# 6/9
### Agenda
- Status
  - [Design doc](https://github.com/pdgetrf/ArktosEdge/blob/main/design/530_design.md)
  - Deadlines:
    - [ONE deadline 6/20](https://events.linuxfoundation.org/open-networking-edge-summit-north-america/program/cfp), ICC scope
    - [OSS deadline 6/13](https://events.linuxfoundation.org/open-source-summit-north-america/), general scope
  - 7/30-> 8/30 Release Planning
    - Items: 
      - Edge Cluster
      - Cascading Clusters
      - Inter-cluster Communication
    - Repo
    - Feature gap sync with Akraino release/demo
    - UWB edge use case, legal assurence
  - MEC
 - POC
   - Inter-cluster Communication


### Note
- Alcor CNI not compatible?
- Inter-cluster communication scope: 
  - Scoped:
    - pods that belong to the same VPC but in different physical clusters to talk to each other 
    - pods that belong to different VPCs in different physical clusters to talk to each other
  - Non-scoped: 
    - pods that belong to different VPCs in the same physical cluster to talk to each other (Mizar)


# 6/16

### Agenda

- [x] [Open Source Summit](https://events.linuxfoundation.org/open-networking-edge-summit-north-america/program/cfp) proposal
- [ ] [ONE deadline 6/20](https://events.linuxfoundation.org/open-networking-edge-summit-north-america/program/cfp), ICC scope
- POC
  - Inter-cluster Communication
      - Scopes: 
        - pods that belong to the **same** VPC but in different physical clusters to talk to each other 
        - pods that belong to **different** VPCs in different physical clusters to talk to each other 
    - high-level architecture
    - P2P solution scoping
      - https://skupper.io/index.html
      - https://cilium.io/blog/2019/03/12/clustermesh
      - https://libp2p.io/
      - service mesh?
    - [ ] Mizar requirement one-pager

### Notes


# 6/23

### Agenda

- [x] [Open Source Summit](https://events.linuxfoundation.org/open-networking-edge-summit-north-america/program/cfp) proposal
- [x] [ONE deadline 6/20](https://events.linuxfoundation.org/open-networking-edge-summit-north-america/program/cfp), ICC scope
- POC
  - Inter-cluster Communication
    - package paths
    - control plane management
    - [ ] Mizar requirement one-pager
- Next monday community meeting
  - libp2p talk
  - UWB use case walk through

### Notes

# 6/30

### Agenda

- Inter-cluster Communication
  - Status
    - Overall design being finalized
      - [ ] Design doc
        - Scope
        - Data flow
        - Control flow  
    - POC Risk: data flow
      - Connectivity
      - Performance
  - POC
    - Data flow
      - Divider redirect, possibly as a Mizar feature in future release
      - Gateway packet packing and transfering
    - Control flow
      - VPC & subnet creation top-down through Mission
      - VPC & subnet global update as a distributed map (slides)
    - [ ] Mizar requirement one-pager
      - Mizar's pod integration with VPC and subnet
      - network control plane
      - Configurable divider for *XDP_REDIRECT*
    - P2P solution survey by Eunju
- External collab
  - libp2p talk
  - UWB use case to be presented in a few weeks



# 7/7

### Agenda

- Inter-cluster Communication
  - Status
    - Overall design being finalized
      - [x] [Design doc](https://github.com/pdgetrf/ArktosEdge/blob/main/design/530_design.md#inter-cluster-communication)
    - POC Risk: data plane flow
      - [tasks](https://github.com/pdgetrf/ArktosEdge/projects/2)
      - Targeting mid-July
      - Daily scrum
      - Need tech support from Mizar team on env setup and debugging
  - POC
    - Data flow
      - Divider redirect, possibly as a Mizar feature in future release
      - Gateway packet packing and transfering
    - Control flow
      - VPC & subnet global synchronization (slides)
    - [x] [Mizar requirement one-pager](https://github.com/CentaurusInfra/mizar/issues/505#issuecomment-875104608)
      - Mizar's pod integration with VPC and subnet
      - Network control plane
      - Configurable divider for *XDP_REDIRECT*

- New Finds
  - KubeEdge Robotics SIG
    - [First SIG](https://www.bilibili.com/video/BV1MM4y1M7vF?share_source=copy_web)

### Notes
- Mizar dev tips and tricks https://github.com/CentaurusInfra/mizar/wiki/Mizar-Developer-Tips-&-Tricks
- Akrano release & demo 8/26


# 7/14

### Agenda

- Inter-cluster Communication
  - Status
    - Overall design being finalized
      - [x] [Design doc](https://github.com/pdgetrf/ArktosEdge/blob/main/design/530_design.md#inter-cluster-communication)
    - POC Risk: data plane flow
      - [tasks](https://github.com/pdgetrf/ArktosEdge/projects/2)
      - Targeting mid-July
  - POC
    - Goal: 
      - Redirect packet to gateway process
      - Unblock gateway development

    - Mizar deeper dive and possible routes
      - ["1 program and 3 maps"](https://github.com/pdgetrf/ArktosEdge/blob/main/slides/how%20does%20Mizar%20maps%20work.pptx)
      - Possible routes for packet redirect at divider
        - ~~Inject gateway ip into the ["network map"](https://github.com/CentaurusInfra/mizar/blob/e8c21f5f262d79dd71cfec5e511a898c7cb1dbe9/src/xdp/trn_transit_xdp_maps.h#L46) directly on host~~
        - Modify operator to inject gateway ip (possible [here](https://github.com/CentaurusInfra/mizar/blob/e8c21f5f262d79dd71cfec5e511a898c7cb1dbe9/mizar/dp/mizar/workflows/dividers/create.py#L59)) (owner: Shaojun)
        - Modify existing subnet0's endpoint value to point to gateway ip (note: currently default subnet0 consumes the entire CIDR space of vpc0, so the goal here is to simply divert packet traffic on divider to external gateway so gateway work can be unblocked) (owner: Qian)
        - Modify transit XDP to redirect (release solution (possibly [here](https://github.com/CentaurusInfra/mizar/blob/e8c21f5f262d79dd71cfec5e511a898c7cb1dbe9/src/xdp/trn_transit_xdp.c#L132))
        
- Collaboration
  - UWB edge requirement disucssion on 15th

### Notes

