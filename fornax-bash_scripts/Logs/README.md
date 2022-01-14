# Fornax Bash Scripts for Automating the Deployment (On-Premise)

### 1. Generate ssh key in Node-B and copy ssh key ID to Node-A & Node-C:
       ssh-keygen -t rsa
       ssh-copy-id (node-a IP)
       ssh-copy-id (node-c IP)

### 2. Clone the repository into Node-A, Node-B, Node-C
       git clone https://github.com/Click2Cloud-Centaurus/fornax.git -b test

### 3. Edit the IP's in all the three scripts before running:
       export a= (IP address of node-a)
       export b= (IP address of node-b)
       export c= (IP address of node-c)

### 4. Run the Scripts:
       sudo bash /root/fornax/12-01-2022/fornax1.sh (for node-a)
       sudo bash /root/fornax/12-01-2022/fornax2.sh (for node-b)  (run the script only after successfully running the node-a script)
       sudo bash /root/fornax/12-01-2022/fornax3.sh (for node-c)  (run the script only after successfully running the node-b script)
  
### 5. Verify the Edgecluster in Node-A:
       kubectl get edgecluster
       
### 6. To see Cloudcore & Edgecore logs:
       cd /root/go/src/github.com/kubeedge
       cat cloudcore.logs
       cat edgecore.logs
