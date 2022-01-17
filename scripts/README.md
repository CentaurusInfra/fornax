# Bash Scripts for Fornax Deployment
### Node-A: Cloud Core Node 
### Node-B: Edge Node with Control Plane 
### Node-C: Edge Worker Node
### 1. Generate ssh key in Edge Node with Control Plane (node-b) and copy ssh key ID to Cloud Core Node (node-a) & Edge Worker node (node-c):
       ssh-keygen
       ssh-copy-id (node-a IP)
       ssh-copy-id (node-c IP)

### 2. Edit the IP's in 'cloud-core.sh ' & 'edge-node-control-plane.sh':
       declare -x a= (IP address of node-a)
       declare -x b= (IP address of node-b)
       declare -x c= (IP address of node-c)

### 3. Run the Scripts:
       sudo bash cloud-core.sh (for node-a)
       sudo bash edge-node-control-plane.sh (for node-b)  (run the script only after successfully running the node-a script)
       sudo bash worker-node.sh (for node-c)  (run the script only after successfully running the node-b script)
  
### 4. Verify the Edgecluster in 'Cloud Core Node' (Node-A):
       kubectl get edgecluster
       
### 5. To see Cloudcore & Edgecore logs:
       cd $HOME/go/src/github.com/fornax
       cat cloudcore.logs
       cat edgecore.logs

