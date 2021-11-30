#	Virtual Machine Setup and Configuration

## Abstract
The purpose of this document is to how to setup and configuration Virtual Machine, and describe the each step to create virtual machine, setup the port number.
1. Virtual Machine Setup (create Cloud core  and Edge core virutal machine, and setup port),

## 1.1. Virtual Machine Setup and Configuration (We use AWS for example)
-	Ubuntu 18.04, one for cloud-core, two for edge-core.
-	Open the port of 10000 and 10002 in the security group of the cloud-core machine and edge-core machine

### 1.1.1. SETUP CLOUD CORE VIRTUAL MACHINE
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

###  1.1.2.	REPEAT 1.1.1 CREATE TWO EDGE-CORE Virtual Machine
•	Create two Edge core in AWS. And Setup port and security, disk space, Unix Ubuntu machine.

###  1.1.3.	After you done above section, you can got 1.2.  <a href="vrital_setup.md" target="_blank"> Install Kubernetes Tools to Cloud core and Edge core </a>
