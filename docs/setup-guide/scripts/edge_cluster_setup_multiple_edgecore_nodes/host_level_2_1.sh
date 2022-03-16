#! /bin/bash

set -e 

echo -e "\n"
#To kill running process of edgecore
edgecore=`ps -aef | grep _output/local/bin/edgecore | grep -v sh| grep -v grep| awk '{print $2}'`
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
if [ -f "${DIR}/prerequisite_packages.sh" ]; then
source "${DIR}/prerequisite_packages.sh"
fi

echo -e "## SETTING UP THE HOSTNAME NODE-B\n"
sudo hostnamectl set-hostname node-b
echo -e "## DISABLING FIREWALL\n"
sudo ufw disable
sudo swapoff -a

key_gen(){
   if [ "$(ls /root/.ssh/id_rsa.pub)" != "/root/.ssh/id_rsa.pub" ] > /dev/null 2>&1
   then
       echo -e "## GENERATING KEY AND COPYING THE KEY TO HOST MACHINE 3"
       chmod 600 $key_pair_3 
       < /dev/zero ssh-keygen -q -N ""
       cat ~/.ssh/id_rsa.pub | ssh -o StrictHostKeyChecking=no -i $key_pair_3 root@$ip_m3 "cat >> ~/.ssh/authorized_keys"
   else
       cat ~/.ssh/id_rsa.pub | ssh -o StrictHostKeyChecking=no -i $key_pair_3 root@$ip_m3 "cat >> ~/.ssh/authorized_keys"
   fi
} 

edgecore_process(){
   if `[ !-z "$edgecore"]`
   then
      echo edgecore process is not running 
   else
      kill -9 $edgecore
      echo edgecore process killed forcefully, process id $edgecore.
   fi
}

fornax_setup_vm_2(){
    echo  -e "## FORNAX CONFIGURATION"
    pushd $HOME/go/src/github.com/fornax
    cp $HOME/machine_1_admin_file/admin.conf $HOME/go/src/github.com/fornax
    systemctl restart docker

    chmod a+x Makefile
    make all
    make WHAT=edgecore
    mkdir /etc/kubeedge/config -p
    cp /etc/kubernetes/admin.conf $HOME/.kube/config

    echo '## SETTING UP THE EDGECORE' 
    cp /etc/kubernetes/admin.conf  $HOME/edgecluster.kubeconfig
    _output/local/bin/edgecore --edgeclusterconfig > /etc/kubeedge/config/edgecore.yaml
    tests/edgecluster/hack/update_edgecore_config.sh admin.conf
    echo '## APPLYING DEVICES.YAML'
    kubectl apply -f build/crds/devices/devices_v1alpha2_device.yaml
    kubectl apply -f build/crds/devices/devices_v1alpha2_devicemodel.yaml
    kubectl apply -f build/crds/reliablesyncs/cluster_objectsync_v1alpha1.yaml
    kubectl apply -f build/crds/reliablesyncs/objectsync_v1alpha1.yaml
    kubectl apply -f build/crds/router/router_v1_rule.yaml
    kubectl apply -f build/crds/router/router_v1_ruleEndpoint.yaml
    kubectl apply -f build/crds/edgecluster/mission_v1.yaml
    kubectl apply -f build/crds/edgecluster/edgecluster_v1.yaml
    chmod 777 $HOME/go/src/github.com/fornax/_output/local/bin/kubectl/vanilla/kubectl
    echo '## RUN EDGECORE'
    export KUBECONFIG=/etc/kubernetes/admin.conf
    nohup _output/local/bin/edgecore --edgecluster >> edgecore.logs 2>&1 &
}
#key_gen

edgecore_process

ip_tables

docker_install

kube_packages

kube_cluster

golang_tools

fornax_setup_vm_2
echo -e "## SETUP SUCCESSSFUL\n"
echo -e "## Logs:"
echo -e "Edgecore: $HOME/go/src/github.com/fornax/edgecore.logs\n"
