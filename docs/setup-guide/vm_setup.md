#	Create, Setup and Configuration Virtual Machine

## Abstract
The purpose of this document is how to create, setup and configuration virtual machine, and describe the each step to create virtual machine, setup the port number.
1. Instruction for create virtual machine (create virtual machine A as cloud core, create virtual machine B and C as edge core).

## 1.1. Create Virtual Machine, and Setup and Configuration (We use AWS for example)
-	Ubuntu 18.04, one for cloud-core, two for edge-core.
-	Open the port of 10000 and 10002 in the security group of the cloud-core machine and edge-core machine
-	**Create virtual machine from brand new instance. See 1.1.1**
-	**Create virtual machine from exit instance. See 1.1.2**

### 1.1.1 Create Clore Core Virtual Machine A From Brand New Instance
- Select instance launch from the right up corner
<img src="images/EC2_01_chooseami_01.png" 
     width="98%"  
     align="center"/>

- Select vertiual machine type: Ubuntu 18.04
<img src="images/EC2_01_chooseami_02.png" 
     width="98%"  
     align="center"/>
     
- Choose Instance Type : t2.large or t2.xlarge
<img src="images/EC2_02_chooseinstancetype.png" 
     width="98%"  
     align="center"/>
     
- Configure Instance: 
<img src="images/EC2_03_configureinstance.png" 
     width="98%"  
     align="center"/>
     
- Add Storage
<img src="images/EC2_04_addstorage.png" 
     width="98%"  
     align="center"/>
     
- Add Tags
<img src="images/EC2_05_addtags.png" 
     width="98%"  
     align="center"/>
     
- Configure Security Group
<img src="images/EC2_06_configsecuritygroup.png" 
     width="98%"  
     align="center"/>
     
- Review
<img src="images/EC2_07_review.png" 
     width="98%"  
     align="center"/>
     
- Select Key pair
<img src="images/EC2_08_selectkeypair.png" 
     width="98%"  
     align="center"/>
     
- Final review and launch
<img src="images/EC2_09_lauchstatus.png" 
     width="98%"  
     align="center"/>

### 1.1.2 Create Cloud Core Virtual Machine From Exist Instance.
-	This Step to create Cloud core in AWS. And Setup port and security, disk space, Unix Ubuntu machine.
-	If you already have similarity machine, you can follow step to create a virtual machine (if you did not have, and you can create brand new from the scratch).
-	In AWS EC2, pickup instance which you want to copy. Then pickup “Launch more like this”.

<img src="images/EC2_vm_01.png" 
     width="98%"  
     align="center"/>
     
<img src="images/EC2_vm_02.png" 
     width="98%"  
     align="center"/>
     
- 	You will get following screen for review.  
<img src="images/EC2_vm_03_review.png" 
     width="98%" 
     align="center"/>
-	Change disksapce size to 80G 
<img src="images/EC2_vm_04_storage.png" 
     width="98%" 
     align="center"/>
- 	Give a Tags name. see screen shot.
<img src="images/EC2_vm_05_tagname.png" 
     width="98%" 
     align="center"/>
- 	Click "Review and Launch" button to review.
<img src="images/EC2_vm_06_security.png" 
     width="98%" 
     align="center"/>
- 	Click "Launch"
<img src="images/EC2_vm_07_launch.png" 
     width="98%" 
     align="center"/>
- 	It will pop up a window, pickup "Choose an existing key pair" and edge-team-key|RSA. Check "checkbox", then click "Launch Instance"
<img src="images/EC2_vm_08_keypair.png" 
     width="98%" 
     align="center"/>
- Waiting virtual machine to launch. The following window will show. Then  click "View Instance"
<img src="images/EC2_vm_09_status.png" 
     width="98%" 
     align="center"/>

<img src="images/EC2_vm_10_view.png" 
     width="98%" 
     align="center"/>
- 	Your instance  will be  running.
<img src="images/EC2_vm_11_cloudcore.png" 
     width="98%" 
     align="center"/>
- 	Add prot number 10000 and 10002.
<img src="images/EC2_vm_12_port1.png" 
     width="98%" 
     align="center"/>

<img src="images/EC2_vm_13_port2.png" 
     width="98%" 
     align="center"/>
- 	Click "Edit inbound rules"
<img src="images/EC2_vm_14_portadd.png" 
     width="98%" 
     align="center"/>

<img src="images/EC2_vm_15_portdone.png" 
     width="98%" 
     align="center"/>
- 	Finally you will see 10000 and 10002 port.
<img src="images/EC2_vm_16_portfinal.png" 
     width="98%" 
     align="center"/>

###  1.1.3	Repeat 1.1.1 Create B and C Edgecore Virtual Machine
•	Create two Edge core virtual machine(B and C) in AWS. And Setup port and security, disk space, Unix Ubuntu machine.

###  1.1.4	After you done above section, go to 1.2  [Install Kubernetes Tools](cluster_setup.md)
