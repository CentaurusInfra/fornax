# Fornax Cluster Multi-Layer: Multi-Edge Node in the same level Setup using Bash Scripts


### Virtual Machine Configuration 

-	**3 Ubuntu 18.04 VMs, one for cloud-core, two for edge-core.**
-	Open the port of 10000 and 10002 in the security group of the cloud-core machine and edge-core machine   
-	EC2 Instance: `t3.2xlarge, 128 GB Storage`.

####    Host Machine 1: Cloud Core Node (Root Host)
####    Host Machine 2: Edge Core Node
####    Host Machine 3: Edge Core Node

### A Step-by-Step Process for Setting up all Machines (For AWS EC2 Instances)

- **Step 1.1:  Switch to ROOT user:**

```bash
sudo su
```
- **Step 1.2: Run the following command on all the Three Host Machines only after switching to 'root' user**

```bash
cat /home/ubuntu/.ssh/authorized_keys  > /root/.ssh/authorized_keys
```

- **Step 1.3: Create project folder and Clone the Fornax Repository**

```bash
mkdir -p /root/go/src/github.com
cd /root/go/src/github.com
git clone https://github.com/CentaurusInfra/fornax.git
```
## Run the scripts (Only after completing step 1. in all the three machines):

### For Host Machine 1:

- **Step 2.1: Create two empty files (like aws-keypair-2.pem & aws-keypair-3.pem) with extension `.pem`  in host-1 & Update these `.pem` files by copying the content of host-2 & host-3 `aws-keypair` `.pem` (keypair which was generated while launching the instance-2 and instance-3 in AWS) files respectively :**
Note: if you use the same keypair pem file when you create aws instance in EC2, you can use the same file name.

```bash
touch aws-keypair-2.pem
vi  aws-keypair-2.pem
```
```bash
touch aws-keypair-3.pem
vi  aws-keypair-3.pem
```
- **Step 2.2: Run the command**
```bash
sudo bash fornax/docs/setup-guide/scripts/host_1.sh
```
- **Step 2.3: Input the Private IP's of Hosts and keypair path:**

   ![image](https://user-images.githubusercontent.com/95343388/154034770-7a8028ee-6ebc-42b7-ae2c-ac254a3f256b.png)
   

### For Host Machine 2: (Run the script only after successfully running the Machine-1 script)

- **Step 3: Run the command**

```bash
sudo bash fornax/docs/setup-guide/scripts/host_2_1.sh
```
 

### For Host Machine 3: (Run the script only after successfully running the Machine-2 script)

- **Step 4: Run the command**

```bash
sudo bash fornax/docs/setup-guide/scripts/host_2_2.sh
```

**Note:  `prerequisite_packages.sh` contains all the required packages for creating Kubernetes Cluster.**


#### Verify the Edge cluster by running command in Host Machine 1:

```bash
kubectl get edgecluster
```
  ![image](https://user-images.githubusercontent.com/95343388/154036219-3314f23a-1828-4598-afa2-9c4cada412c7.png) 


#### To see Cloudcore & Edgecore logs:
```bash
cd $HOME/go/src/github.com/fornax
cat cloudcore.logs
cat edgecore.logs
```

#### To verify Cloudcore & Edgecore is running currently:
```bash
ps -aef | grep _output/local/bin/cloudcore | grep -v sh| grep -v grep
ps -aef | grep _output/local/bin/edgecore | grep -v sh| grep -v grep
cat 