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

# 7/21

### Agenda
- Overall status
  - OSS not accepted :(
    - Priority: 
```
    design --> release -> perf test 
                  |
                  +-> academia conferences -> exposure conferences
```
  - Project tracked in two fronts:
    - Release
      - [x] Repo & CICD (owner: Qian)
      - [ ] Edge Cluster featrues (owner: Qian)
    - POC: Inter-cluster communication  
- Inter-cluster Communication status
  - POC Risk: data plane flow
    - [tasks](https://github.com/pdgetrf/ArktosEdge/projects/2)
    - Targeting mid-July
  - POC
    - Goal: 
      - Redirect packet to gateway process
      - Unblock gateway development
    - More Mizar deep dive and possible routes
      - Possible routes for packet redirect at divider
        - ~~[ ] Inject gateway ip into the ["network map"](https://github.com/CentaurusInfra/mizar/blob/e8c21f5f262d79dd71cfec5e511a898c7cb1dbe9/src/xdp/trn_transit_xdp_maps.h#L46) directly on host~~
        - [x] Modify operator to inject gateway ip (possible [here](https://github.com/CentaurusInfra/mizar/blob/e8c21f5f262d79dd71cfec5e511a898c7cb1dbe9/mizar/dp/mizar/workflows/dividers/create.py#L59)) (owner: Shaojun)
        - ~~[ ] Modify existing subnet0's endpoint value to point to gateway ip (note: currently default subnet0 consumes the entire CIDR space of vpc0, so the goal here is to simply divert packet traffic on divider to external gateway so gateway work can be unblocked)~~
      - Release solution
        - Modify transit XDP to redirect (release solution (possibly [here](https://github.com/CentaurusInfra/mizar/blob/e8c21f5f262d79dd71cfec5e511a898c7cb1dbe9/src/xdp/trn_transit_xdp.c#L132))
        - Mizar control plane update the network map based on gateway's map
        
- Collaboration
  - UWB to present next week


# 8/3

### Agenda
- Priority
```
    design --> release -> perf test 
                  |
                  +-> academia conferences -> exposure conferences
```

- Overall status
  - Project tracks:
    - Release
      - [x] [Release Plan](https://github.com/CentaurusInfra/fornax/blob/main/docs/fornax-design/release_plan.md)
      - [x] Repo & CICD (owner: Qian)
      - [x] Edge Cluster featrues (owner: Qian)
    - POC: Inter-cluster communication  
      - Goal: 
        - Packet e2e
        - Detailed gateway design
    - [POD to VPC work by Mizar team](https://github.com/CentaurusInfra/mizar/pull/518)
  - Exposures
    - ONE & KubeCon NA not accepted :(
      - Reasons
    - Peer-reviewed Academic Conferences
      - CFP 
        - [IEEE Edge 2021](https://conferences.computer.org/edge/2021/cfp/), deadline **9/15**
        - [The Sixth International Conference on Fog and Mobile Edge Computing](https://emergingtechnet.org/FMEC2021/)
      - Focuses
        - "Releated work"
        - Application scenarios

- Collaboration
  - UWB stroke recovery project

- Akraino Release (8/26)


# 8/11

### Agenda
- Overall status
  - Project tracks:
    - Release
      - [x] [Release Plan](https://github.com/CentaurusInfra/fornax/blob/main/docs/fornax-design/release_plan.md)
      - [x] Edge cluster featrues (owner: Qian)
        - [ ] [Detailed test plan](https://github.com/CentaurusInfra/fornax/blob/release-test-plan/docs/fornax_test/830_release_testplan.md)
          - Manual at the moment, automated after 8/30 release
          - ~20 key scenarios (more to come) 
        - [ ] Edge application on hierarchical clusters
          - Goal: demo the benefits of hierarchical edge clusters 
          - Latency
          - Autonomy
          - Distributed cloud (offload cloud application to the edge, together with dependencies)
          - 5G (network speed > local disk I/O --> short video encoding/decoding, gaming)
      - Inter-cluster communication 8/30 
        - release components: Gateway
          1. Design doc + E2E POC Demo
          2. Initial version of the Gateway, tested against Mizar's Kind env + Edge cluster release env
    - POC: Inter-cluster communication  
      - Goal: 
        - Pod traffic e2e
        - Detailed gateway design
      - Tech discussion
        - Multi-cluster VPC/Subnet
        - VNI translation between different clusters
        - Direct-path

- Akraino Release (8/26)


# 8/18

### Agenda
- Overall status
  - Project tracks:
    - Release
      - [x] [Release Plan](https://github.com/CentaurusInfra/fornax/blob/main/docs/fornax-design/release_plan.md)
      - [x] Edge cluster featrues (owner: Qian)
        - [ ] (Merged) [Detailed test plan](https://github.com/CentaurusInfra/fornax/blob/main/docs/fornax_test/830_release_testplan.md)
        - [ ] Edge application on hierarchical clusters (owner: Qian)
          - Goal: demo the benefits of hierarchical edge clusters 
          - Benchmark
          - CDN
      - Inter-cluster communication 8/30 
        - Release components: Gateway
          1. Design doc + E2E POC Demo
          2. Initial version of the Gateway, tested against Mizar's Kind env + Edge cluster release env
        - POC
          - Goal: 
            - Pod traffic e2e
            - Detailed gateway design
          - Tech discussion
            - Direct-path across clusters
  - Exposures
    - Peer-reviewed Academic Conferences
      - CFP 
        - [IEEE Edge 2021](https://conferences.computer.org/edge/2021/cfp/), deadline **9/15**
        - [The Sixth International Conference on Fog and Mobile Edge Computing](https://emergingtechnet.org/FMEC2021/)
      - Focuses
        - "Releated work"
        - Application scenarios

- Akraino Release (8/26)


# 8/25

### Agenda
- Overall status
  - Project tracks:
    - Release
      - Adjustment
        - 830 release with edge cluster
          - Cut release branch on 8/31 (next Wednesday)
        - inter-cluster communication to be released in 930 (together with any edge cluster fixes)
      - Edge application on hierarchical clusters (owner: Qian)
          - Goal: demo the benefits of hierarchical edge clusters 
          - Idea 1: Edge Benchmark
            - Represents a class of applications that fit a certain profile. In terms of edge, 
              - Latency & data locality
                - Large data volume on the edge
                - Data has regional features
              - Local vs global processing
                - Edge (Local) processing + global aggregation
              - Autonomous against resource event
                - Network failure
                - Node failure
              - Remote management
                - Deploy and manage from upper level
          - Idea 2: Real-world application
          - 9/27 OSS edge talk with link to demo video
      - Inter-cluster communication 9/30 
        - POD VPC PR from Phu
        - Switching from Kind env to K8s/Arktos env
        - POC
          - Goal: 
            - Pod traffic e2e
            - Detailed gateway design
  - Exposures
    - Peer-reviewed Academic Conferences
      - CFP 
        - [IEEE Edge 2021](https://conferences.computer.org/edge/2021/cfp/), deadline **9/15**
        - [The Sixth International Conference on Fog and Mobile Edge Computing](https://emergingtechnet.org/FMEC2021/)
      - Focuses
        - "Releated work"
        - Application scenarios

- Akraino Release


# 9/1

### Agenda
- Overall status
  - Project tracks:
    - Release
      - [x] 830 release
        - [Release cut](https://github.com/CentaurusInfra/fornax/releases/tag/v0.1) 
      - Edge application on hierarchical clusters (owner: Qian)
          - Goal: demo the benefits of hierarchical edge clusters 
            - AI application demo (by Qian)
            - Benchmarking
            - 9/27 OSS edge talk with link to demo video (todo)
            - Serverless
      - Inter-cluster communication 9/30 
        - Switching from Kind env to real K8s cluster env (owner: Peng)
          - [x] Env setup
          - [ ] Documentation 
        - POC
          - Goal: 
            - Pod traffic e2e
            - Detailed gateway design
      - New release brainstorming
        - 5G application
  - Exposures
    - Kubecon Publication, topic "Edge Networking with Mizar", 600-1000 words, due by 9/10 (owner Peng)
    - Friday brown bag talks
    - Peer-reviewed Academic Conferences
      - CFP 
        - [IEEE Edge 2021](https://conferences.computer.org/edge/2021/cfp/), deadline **9/15** (not likely to make it)
        - [NSDI](https://www.usenix.org/conference/nsdi22/call-for-papers), Paper titles and abstracts due: 9/9, full paper 9/15
        - [The Sixth International Conference on Fog and Mobile Edge Computing](https://emergingtechnet.org/FMEC2021/) (need to look into it)
- Akraino Release
  - On track



# 9/8

### Agenda
- Overall status
  - Project tracks:
    - Release
      - [x] [830 release](https://github.com/CentaurusInfra/fornax/releases/tag/v0.1) 
      - [ ] 930 release "Inter-cluster communication"
        - Edge Cluster Mission Improvement inspired by AI demo experience (owner: Qian)
          - Goal: to allow deployment of AI demo completely with Mission from cloud to edge
        - Switching from Kind env to real K8s cluster env (owner: Peng)
          - [x] Env setup (quick demo)
          - [x] [Documentation](https://github.com/pdgetrf/mizar_cluster_scripts) 
        - POC
          - Goals (targeting mid Sept)
            - Pod traffic e2e (to use proxy as gateway)
            - Detailed gateway design
        - Release item (some risk due to team size change, to be further estimated by end of this week)
          - Gateway
      - New release brainstorming (next week after 15th)
  - Exposures
    - [x] Friday brown bag talk, [video](https://www.youtube.com/watch?v=W0egc5W3Q2Q)
    - [ ] Kubecon Publication, topic "Edge Networking with Mizar", 600-1000 words, due by 9/10 (owner Peng)
    - Peer-reviewed Academic Conferences
      - CFP 
        - [NSDI](https://www.usenix.org/conference/nsdi22/call-for-papers)
          - [ ] Paper titles and abstracts due: 9/9
          - [ ] full paper 9/15
- "Macro trend" discussion (due by 10/31)
  - 5G application
  - Cloud-cloud vs Cloud-edge
- Akraino Release
  - On track


# 9/15

### Agenda
- Overall status
  - Project tracks:
    - Release
      - [x] [830 release](https://github.com/CentaurusInfra/fornax/releases/tag/v0.1) 
      - [ ] 930 release "Inter-cluster communication"
        - [x] Edge Cluster Mission Improvement inspired by AI demo experience (owner: Qian)
        - POC
          - Goals (targeting mid Sept)
            - Pod traffic e2e (to use proxy as gateway)
            - Detailed gateway design
        - Release item (to be further estimated by end of this week)
      - New release brainstorming (next week after 15th)
  - Exposures
    - [x] [Kubecon Publication](https://vmblog.com/archive/2021/09/13/towards-a-scalable-reliable-and-secure-edge-computing-framework.aspx#.YUJra55Kg8N)
    - Peer-reviewed Academic Conferences
      - CFP 
        - [NSDI](https://www.usenix.org/conference/nsdi22/call-for-papers), will aim for their next April deadline due to team size change
- "Macro trend" discussion (due by 10/31)
  - 5G application
  - Cloud-cloud vs Cloud-edge
- Akraino Release
  - On track (documentation owner?)


# 9/21

### Agenda
- Overall status
  - Project tracks:
    - Release
      - [ ] 930 release "Inter-cluster communication"
        - [x] Edge Cluster Mission Improvement inspired by AI demo experience (owner: Qian)
          - "One-click" edge app deployment
          - Front-end needs more work for a better live demo
          - Team demo 
        - Edge-edge communication
          - Automated dev env setup for Qian and David
          - POC/Dev
            - Pod traffic e2e (to use proxy as gateway) (70%)
              - [x] Mizar control plane changes for gateway host
              - [x] Mizar data plane to 
                  1. route traffic to user space on gateway host
                  2. avoid ep_host_cache on divider when traffic comes from the gateway host
              - [ ] connect 2nd cluster and perform e2e test
            - Detailed gateway design
        - Release items
      - New release brainstorming (next week after 27th)
  - Exposures
    - [x] [Kubecon Publication](https://vmblog.com/archive/2021/09/13/towards-a-scalable-reliable-and-secure-edge-computing-framework.aspx#.YUJra55Kg8N)
    - [ ] OSS tutorial talk 9/27
- "Macro trend" discussion (due by 10/31)
  - 5G application
  - Cloud-cloud vs Cloud-edge
  - Paper reading
    - Deep Learning With Edge Computing- A Review
    - Adaptive Federated Learning in Resource Constrained Edge Computing Systems
    - Edge Intelligence- Paving the Last Mile of Artificial Intelligence With Edge Computing
    - The Emerging Landscape of Edge-Computing
- Akraino Release
  - documentation owner settled

# 9/29

### Agenda
- Overall status
  - Project tracks:
    - Release
      - [ ] 930 release "Inter-cluster communication"
        - Edge-edge communication
          - Automated dev env setup for Qian and David
          - POC/Dev
            - Pod traffic e2e (to use proxy as gateway) (70%)
              - [x] Mizar control plane changes for gateway host
              - [x] Mizar data plane to 
                  1. [x] route traffic to user space on gateway host
                  2. [x] avoid ep_host_cache on divider when traffic comes from the gateway host
              - [ ] connect 2nd cluster and perform e2e test
            - Detailed gateway design, 930 release
      - New release brainstorming (next week after 27th)
  - Exposures
    - [x] OSS tutorial talk 9/27
- "Macro trend" discussion (due by 10/31)
  - 5G application
  - Cloud-cloud vs Cloud-edge
  - Serverless
  - Paper reading
    - Deep Learning With Edge Computing- A Review
    - Adaptive Federated Learning in Resource Constrained Edge Computing Systems
    - Edge Intelligence- Paving the Last Mile of Artificial Intelligence With Edge Computing
    - The Emerging Landscape of Edge-Computing
- Onboarding
- Akraino Release
  - documentation owner settled
- Notes:
  - [Edge gallery for 5G](https://gitee.com/organizations/edgegallery/projects)


# 10/06

### Agenda
- Overall status
  - Project tracks:
    - Release
      - [x] 930 release
        - Edge-edge communication
      - New release brainstorming (next week)
  - Exposures
    - [x] KubeCon virtual booth slides tutorial talk 9/27
- "Macro trend" discussion (due by 10/31)
  - 5G application
  - Cloud-cloud vs Cloud-edge
  - Serverless
  - [Edge gallery for 5G MEC](https://gitee.com/organizations/edgegallery/projects)
  - Paper reading
    - [ ] Deep Learning With Edge Computing- A Review
    - [ ] Adaptive Federated Learning in Resource Constrained Edge Computing Systems
    - [ ] Edge Intelligence- Paving the Last Mile of Artificial Intelligence With Edge Computing
    - [ ] The Emerging Landscape of Edge-Computing
- [x] Onboarding
- Akraino Release
  - [ ] documentation owner settled
