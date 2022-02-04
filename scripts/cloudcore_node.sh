#! /bin/bash

set -e

# Enter the IP address of the a: Cloud Core Node, b: Edge Node with Control Plane, c: Edge Worker node 
echo "Enter IP ADDRESS of Machine 1:"
read ip_m1 
echo "Enter IP ADDRESS of Machine 2:"
read ip_m2
echo "Enter ROOT password of Machine 2 for copying the CA, CERTS and Kubeconfig files from Machine 1"
read -s pass_2
echo "Enter IP ADDRESS of Machine 3:"
read ip_m3
echo "Enter ROOT password of Machine 3 for copying the CA, CERTS files to Machine 3"
read -s pass_3
#To kill running process of cloudcore
cloudcore=`ps -aef | grep _output/local/bin/cloudcore | grep -v sh| grep -v grep| awk '{print $2}'`
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
if [ -f "${DIR}/prerequisite_packages.sh" ]; then
source "${DIR}/prerequisite_packages.sh"
fi

key_gen (){
   rm -rf /root/.ssh/id_rsa  &&  rm -rf /root/.ssh/id_rsa.pub
   echo y | apt-get update
   echo y | apt-get install sshpass
   < /dev/zero ssh-keygen -q -N ""
   sshpass -p $pass_2 ssh-copy-id -o StrictHostKeyChecking=no root@$ip_m2
   sshpass -p $pass_3 ssh-copy-id -o StrictHostKeyChecking=no root@$ip_m3
}

pushd $HOME
echo '*****SETTING UP THE HOSTNAME NODE-A*****'
sudo hostnamectl set-hostname node-a
echo '*****DISABLING FIREWALL*****'
sudo ufw disable
sudo swapoff -a

cloudcore_process(){
   if `[ !-z "$cloudcore"]`
   then
      echo cloudcore process is not running 
   else
      kill -9 $cloudcore
      echo cloudcore process killed forcefully, process id $cloudcore.
   fi
}

fornax_setup_vm_1(){
    echo '*****FORNAX CONFIGURATION*****'
    pushd $HOME/go/src/github.com/fornax
    sudo rm -rf ca certs /etc/kubeedge/
    chmod a+x Makefile
    make all
    make WHAT=cloudcore
    mkdir /etc/kubeedge/config -p
    cp /etc/kubernetes/admin.conf $HOME/.kube/config
    _output/local/bin/cloudcore --minconfig > /etc/kubeedge/config/cloudcore.yaml
    sed -i 's+RANDFILE+#RANDFILE+g' /etc/ssl/openssl.cnf
    mkdir -p /etc/kubeedge/ca
    mkdir -p /etc/kubeedge/certs
    build/tools/certgen.sh genCA $ip_m1 $ip_m2 $ip_m3
    build/tools/certgen.sh genCertAndKey server $ip_m1 $ip_m2 $ip_m3
    ssh -t root@$ip_m2 "mkdir -p /etc/kubeedge" > /dev/null 2>&1
    scp -r /etc/kubeedge/certs  $ip_m2:/etc/kubeedge
    scp -r /etc/kubeedge/ca  $ip_m2:/etc/kubeedge
    ssh -t root@$ip_m3 "mkdir -p /etc/kubeedge" > /dev/null 2>&1
    scp -r /etc/kubeedge/certs  $ip_m3:/etc/kubeedge
    scp -r /etc/kubeedge/ca  $ip_m3:/etc/kubeedge
    ssh -t root@$ip_m2 "mkdir -p $HOME/machine_1_admin_file" > /dev/null 2>&1
    scp -r /etc/kubernetes/admin.conf  $ip_m2:$HOME/machine_1_admin_file
    kubectl apply -f build/crds/devices/devices_v1alpha2_device.yaml
    kubectl apply -f build/crds/devices/devices_v1alpha2_devicemodel.yaml
    kubectl apply -f build/crds/reliablesyncs/cluster_objectsync_v1alpha1.yaml
    kubectl apply -f build/crds/reliablesyncs/objectsync_v1alpha1.yaml
    kubectl apply -f build/crds/router/router_v1_rule.yaml
    kubectl apply -f build/crds/router/router_v1_ruleEndpoint.yaml
    kubectl apply -f build/crds/edgecluster/mission_v1.yaml
    kubectl apply -f build/crds/edgecluster/edgecluster_v1.yaml
    export KUBECONFIG=/etc/kubernetes/admin.conf
    nohup _output/local/bin/cloudcore >> cloudcore.logs 2>&1 & 
}
key_gen

cloudcore_process

ip_tables

docker_install

kube_packages

kube_cluster

golang_tools

fornax_setup_vm_1
echo '*****SETUP SUCCESSSFUL*****' 
echo 'Logs: '
echo 'Cloudcore: $HOME/go/src/github.com/fornax/cloudcore.logs'
