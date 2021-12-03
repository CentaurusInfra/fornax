#	File or Folder Copy Process Between Edge cluster For Reference

## Abstract
The purpose of this document is to how to copy file or folder in machine A to machine B and C Virtual Machine(or Cluster), and describe the each step to create copy process and refence documentation.
If you have own way to copy, you can use your own way to copy and skip this doc.
1. Reference document at this link: ,
2. Generate public key in machine A if you want to copy A file or folder to machine B.
3. Copy the public key in A to machine B.
4. Do copy process.

## 1.1. Generate Key in machine A and copy pub key to machne B 
-	Ubuntu 18.04, one for cloud-core, two for edge-core.

### 1.1.1. Generate Key in root at machine A
-	Run following command and keep "Enter" key, until key generated.
```
ssh-keygen -t rsa
```
Result:

### 1.1.2. Open pub key in B.
```
cat .ssh/id_rsa.pub
```
or
```
vi .ssh/id_rsa.pub
```
Result:

### 1.1.3. Copy pub key to Machine B.
Connect to machine B and opend .ssh/authorized_keys
```
vi .ssh/authorized_keys
```
you need append machine A public key to authorized_keys.
Result:


<img src="images/EC2_vm_01.png" 
     width="98%"  
     align="center"/>
