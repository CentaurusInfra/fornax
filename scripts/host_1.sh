#! /bin/bash

set -e

# Enter the IP address of the Host 1, Host 2, Host 3 
echo -e "## Enter Private IP ADDRESS of Host Machine 1:"
read ip_m1 
echo -e "\n"
echo -e "## Enter Private IP ADDRESS of Host Machine 2:"
read ip_m2
echo -e "## Enter Absolute path of your key-pair of Host Machine-2:"
read  key_pair_2 
echo -e "\n"
echo -e "## Enter Private IP ADDRESS of Host Machine 3:"
read ip_m3
echo -e "## Enter Absolute path of your key-pair of Host Machine-3:"
read  key_pair_3
echo -e "\n"

#To kill running process of cloudcore
cloudcore=`ps -aef | grep _output/local/bin/cloudcore | grep -v sh| grep -v grep| awk '{print $2}'`
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
if [ -f "${DIR}/prerequisite_packages.sh" ]; then
source "${DIR}/prerequisite_packages.sh"
fi

pushd $HOME
echo -e "## SETTING UP THE HOSTNAME NODE-A\n"
sudo hostnamectl set-hostname node-a
echo -e "## DISABLING FIREWALL\n"
sudo ufw disable
sudo swapoff -a

key_gen (){
   echo -e "## GENERATING KEY AND COPYING THE KEY TO HOST 2 AND HOST 3"
   if [ "$(ls /root/.ssh/id_rsa.pub)" != "/root/.ssh/id_rsa.pub" ] > /dev/null 2>&1
   then
      < /dev/zero ssh-keygen -q -N ""
      chmod 600 $key_pair_2 && chmod 600 $key_pair_3
      cat ~/.ssh/id_rsa.pub | ssh -o StrictHostKeyChecking=no -i $key_pair_2 root@$ip_m2 "cat >> ~/.ssh/authorized_keys"
      cat ~/.ssh/id_rsa.pub | ssh -o StrictHostKeyChecking=no -i $key_pair_3 root@$ip_m3 "cat >> ~/.ssh/authorized_keys"
   else 
      cat ~/.ssh/id_rsa.pub | ssh -o StrictHostKeyChecking=no -i $key_pair_2 root@$ip_m2 "cat >> ~/.ssh/authorized_keys"
      cat ~/.ssh/id_rsa.pub | ssh -o StrictHostKeyChecking=no -i $key_pair_3 root@$ip_m3 "cat >> ~/.ssh/authorized_keys"
   fi
   echo -e "## KEY GENERATED SUCCESSFULLY\n"
}  

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
    echo -e "## FORNAX CONFIGURATION"
    pushd $HOME/go/src/github.com/fornax
    sudo rm -rf ca certs /etc/kubeedge/
    echo '## SETTING UP THE CLOUDCORE'
    chmod a+x Makefile
    make all
    make WHAT=cloudcore
    mkdir /etc/kubeedge/config -p
    sed -i 's+RANDFILE+#RANDFILE+g' /etc/ssl/openssl.cnf
    cp /etc/kubernetes/admin.conf $HOME/.kube/config
    _output/local/bin/cloudcore --minconfig > /etc/kubeedge/config/cloudcore.yaml
    mkdir -p /etc/kubeedge/ca
    mkdir -p /etc/kubeedge/certs
    build/tools/certgen.sh genCA $ip_m1 $ip_m2 $ip_m3
    build/tools/certgen.sh genCertAndKey server $ip_m1 $ip_m2 $ip_m3
    echo "## COPYING THE KUBECONFIG FILE, CA AND CERTS TO HOST MACHINE 2 AND 3"
    ssh -t root@$ip_m2 "sudo mkdir -p /etc/kubeedge" > /dev/null 2>&1
    sudo scp -r /etc/kubeedge/certs  $ip_m2:/etc/kubeedge
    sudo scp -r /etc/kubeedge/ca  $ip_m2:/etc/kubeedge
    ssh -t root@$ip_m3 "sudo mkdir -p /etc/kubeedge" > /dev/null 2>&1
    sudo scp -r /etc/kubeedge/certs  $ip_m3:/etc/kubeedge
    sudo scp -r /etc/kubeedge/ca  $ip_m3:/etc/kubeedge
    ssh -t root@$ip_m2 "sudo mkdir -p /root/machine_1_admin_file" > /dev/null 2>&1
    sudo scp -r /etc/kubernetes/admin.conf  $ip_m2:/root/machine_1_admin_file
    echo -e "## APPLYING DEVICES.YAML"
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
echo -e "## SETUP SUCCESSSFUL\n" 
echo -e "## Cloudcore Logs: $HOME/go/src/github.com/fornax/cloudcore.logs\n"
