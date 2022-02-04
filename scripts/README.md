# Edge Cluster Multi-Layer Setup using Bash Scripts  



### Virtual Machine Configuration 



•	**3 Ubuntu 18.04 VMs, one for cloud-core, two for edge-core.**   
•	Open the port of 10000 and 10002 in the security group of the cloud-core machine and edge-core machine   
•	16 GB RAM, 16 vCPUs, 128 GB storage.    

####     Machine 1: Cloud Core Node 
####     Machine 2: Edge Node with Control Plane 
####     Machine 3: Edge Worker Node


## Steps to configure 'sshd' before running the scripts in all the three machines:

**Switch to ROOT user:**
        
        sudo -i
        
**Edit the 'sshd_config' file:**

       vi /etc/ssh/sshd_config
       
       
**Here modify line no. 32 and line no. 56 by uncommenting and updating to `PermitRootLogin yes` and `PasswordAuthentication yes`**



   ![image](https://user-images.githubusercontent.com/95343388/152476470-8fb9d893-23bb-4666-84fc-7996f6d132a7.png)
   
   
   

**Now reload the sshd service:**
     
     
       systemctl reload sshd
       
       
**Set the ROOT password of the Machine**


       passwd root
       
       
   ![image](https://user-images.githubusercontent.com/95343388/152478338-2bc2a7da-b236-4776-9c50-42c9eb60eaaf.png)


   
### Running the Scripts:


**create project folder and go to the project folder**

       mkdir -p /root/go/src/github.com
       cd /root/go/src/github.com
       
### Clone the git repo in project folder and run the scripts:


       sudo bash fornax/scripts/cloudcore_node.sh                  (Run in machine-1)
       sudo bash fornax/scripts/edgecore_control_plane.sh          (Run in machine 2)  (run the script only after successfully running the machine-1 script)
       sudo bash fornax/scripts/edge_worker_node.sh                (Run in machine 3)  (run the script only after successfully running the machine-2 script)
       
       
### • Run the machine-2 script only after successfully running the machine-1 script.
### • Run the machine-3 script only after successfully running the machine-2 script.


### NOTE: 'prerequisite_package.sh' contains all the required packages for creating Kubernetes Cluster.
          


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
       
       
       
### If Kubernetes node does not get Ready even after successfully running the script in Machine-3 please run the following command:


       export KUBECONFIG=/etc/kubernetes/admin.conf
       
       
       
   ![image](https://user-images.githubusercontent.com/95343388/152477536-b2aa6c4b-15c5-4b57-87dd-de197d0597c3.png)


