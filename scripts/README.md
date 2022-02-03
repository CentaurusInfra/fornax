# Edge Cluster Multi-Layer Setup using Bash Scripts  



### Virtual Machine Configuration 



•	**3 Ubuntu 18.04 VMs, one for cloud-core, two for edge-core.**   
•	Open the port of 10000 and 10002 in the security group of the cloud-core machine and edge-core machine   
•	16 GB RAM, 16 vCPUs, 128 GB storage.    

####     Machine 1: Cloud Core Node 
####     Machine 2: Edge Node with Control Plane 
####     Machine 3: Edge Worker Node

   
### NOTE: 'prerequisite_package.sh' contains all the required packages for creating Kubernetes Cluster.


   
#### Run the Scripts:


       sudo bash cloudcore_node.sh                  (Run in machine-1)
       sudo bash edgecore_control_plane.sh          (Run in machine 2)  (run the script only after successfully running the machine-1 script)
       sudo bash edge_worker_node.sh                (Run in machine 3)  (run the script only after successfully running the machine-2 script)


#### Input the Private IP's and Password of Machine 1, Machine 2 and Machine 3 :


 **For Machine 1**
       
   ![image](https://user-images.githubusercontent.com/95343388/152158030-2d2a26e9-71e9-4abd-8f04-0330424a32f6.png)

   
 **For Machine 2**
 
 
   ![image](https://user-images.githubusercontent.com/95343388/152291760-fffbe61f-3158-4f3f-b225-e805c608849c.png)

   

#### Verify the Edge cluster by running command in 'Cloud Core Node' (Machine-1):


       kubectl get edgecluster
       
       
       
  ![image](https://user-images.githubusercontent.com/95343388/152162045-d6143680-14eb-470c-89c6-6f4a21e54414.png)

           
           
           
#### To see Cloudcore & Edgecore logs:

       cd $HOME/go/src/github.com/fornax
       cat cloudcore.logs
       cat edgecore.logs
