# Adding Cross Cluster Gateways to Mizar

A gateway is a field belonging to a mizar subnet object that all the traffics go to the gateway if there are any cross cluster requests. 

```
- name: Gateway
  type: string
  priority: 0
  JSONPath: .spec.gateway
  description: The IP of the gateway
```

Also, we add another field External to the subnet object that indicates wether the subnet belongs to the local cluster

```
- name: External
  type: boolean
  priority: 0
  JSONPath: .spec.external
  description: true means that the subnet does not belong to the local cluster; false means that the subnet belongs to the local cluster
```

## Control Plane
For mizar control planes, we add the two fields above to subnet objects. The default value of gateway is empty and the one of external is false. If all the requests happen within the clusters, the two fields do not invovle at all.
For example, there are two clusters A and B. The subnet A works for cluster A as local subnet while subnet B works for cluster B as external subnet. The yaml file applied to cluster A list as below
```
apiVersion: mizar.com/v1
kind: Subnet
metadata:
  name: A
spec:
  bouncers: 1
  ip: "192.168.0.0"
  prefix: "16"
  vpc: "vpc2"
  gateway: ""
  external: false
  status: "Init"
---
apiVersion: mizar.com/v1
kind: Subnet
metadata:
  name: B
spec:
  bouncers: 1
  ip: "192.168.122.0"
  prefix: "16"
  vpc: "vpc2"
  gateway: "172.31.2.217"
  external: true
  status: "Init"
```  
The yaml file applied to cluster B will be changed to
```
apiVersion: mizar.com/v1
kind: Subnet
metadata:
  name: A
spec:
  bouncers: 1
  ip: "192.168.0.0"
  prefix: "16"
  vpc: "vpc2"
  gateway: "172.31.2.33"
  external: true
  status: "Init"
---
apiVersion: mizar.com/v1
kind: Subnet
metadata:
  name: B
spec:
  bouncers: 1
  ip: "192.168.122.0"
  prefix: "16"
  vpc: "vpc2"
  gateway: ""
  external: false
  status: "Init"
```  

### Set up an external subnet with gateway
1. Found a host to deploy gateway 
2. Update the host ip above to replace the gateway value in the given yaml files below. 
3. Verify if the gateway is assigned to the given external subnet bouncer(It has to be 1 bouncer?) 
The code change is at https://github.com/pdgetrf/edge_gateway/blob/71aae9d5de19677a2721eb64db11a3fbb876a509/mizar/dp/mizar/operators/droplets/droplets_operator.py#L99
The external bouncer check is at https://github.com/pdgetrf/edge_gateway/blob/71aae9d5de19677a2721eb64db11a3fbb876a509/mizar/dp/mizar/operators/droplets/droplets_operator.py#L104

### Set up a local subnet(Suppose we create all the external subnets first so that we can exclude gateway host ips from droplets)
1. Get all the subnets
2. Exclude all the gateway host ips from the droplets
3. Verify that the gateway host ips are not asigned to the given local subnet bouncers
The code change is at https://github.com/pdgetrf/edge_gateway/blob/71aae9d5de19677a2721eb64db11a3fbb876a509/mizar/dp/mizar/operators/droplets/droplets_operator.py#L102
4. Verify that the gateway host ips are not asigned to the given local subnet dividers
The code change is at https://github.com/pdgetrf/edge_gateway/blob/71aae9d5de19677a2721eb64db11a3fbb876a509/mizar/dp/mizar/operators/droplets/droplets_operator.py#L129

### Redirect traffic to gateway
1. Suppose the local subnet is 192.168.0.0/16 while the external subnet is 192.168.122.0/16. 
2. For the target ip belonging to the external subnet 192.168.122.0/16, the net data model at https://github.com/pdgetrf/edge_gateway/blob/71aae9d5de19677a2721eb64db11a3fbb876a509/src/xdp/trn_transit_xdp.c#L157 can pass the gateway ip
3. For the change at https://github.com/pdgetrf/edge_gateway/blob/71aae9d5de19677a2721eb64db11a3fbb876a509/src/xdp/trn_transit_xdp.c#L673, a gateway metadata will be added to handle it. 

