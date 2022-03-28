# Fornax cluster scripts to add worker nodes and deploy mizar


### Update existing masters

- **Step 1.1:  Switch to ROOT user:**

```bash
sudo su
```

- **Step 1.2: Go to the deployment script directory and upgrate kernel for mizar**

```bash
cd /root/go/src/github.com/fornax/scripts/deployment
./update_kernel.sh
```

- **Step 1.3: Wait util the master to restart and update the master instance**

```bash
./update_master.sh
```

### Create worker nodes
-	**2 Ubuntu 18.04 VMs**
-	Open the port of 10000 and 10002 in the security group
-	EC2 Instance: `t2.medium, 50 GB Storage`.

- **Step 2.1:  Switch to ROOT user:**

```bash
sudo su
```

- **Step 2.2: Run the script to set up worker nodes**

```bash
cd /root/go/src/github.com/fornax/scripts/deployment
./create_node.sh
```

### Build mizar images and deploy to clusters

- **Step 3.1: Set up hosts for worker nodes and cluster gateway**

- make sure you can ssh from master to all worker nodes  without password. follow [this](http://www.linuxproblem.org/art_9.html) if not. 
- put IPs of worker nodes in a file called **slave.in**. for example:
```bash
cd /root/go/src/github.com/fornax/scripts/deployment
cat slaves.in
18.237.157.249
18.237.205.30
```
- update cluster_gateway.properties by using gateway host ip
```bash
cat ./cluster_gateway.properties 
172.31.15.208
```

- **Step 3.2: Build images**

```bash
./build_mizar_images.sh
```

- **Step 3.3: Restart the cluster**

```bash
./restart_cluster.sh
```

