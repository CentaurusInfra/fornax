#! /bin/bash

set -e 

#Enter kubernetes version

v='1.21.1-00'

#To kill running process of edgecore
edgecore=`ps -aef | grep _output/local/bin/edgecore | grep -v sh| grep -v grep| awk '{print $2}'`

    pushd $HOME
#---------------------------------------------------------------------
echo '*****SETTING UP THE HOSTNAME NODE-C*****'
    sudo hostnamectl set-hostname node-c
#---------------------------------------------------------------------
echo '*****DISABLING FIREWALL*****'
    sudo ufw disable
    sudo swapoff -a
#---------------------------------------------------------------------
ip_tables(){
    echo '*****LETTING IPTABLES SEE BRIDGED TRAFFIC*****'
    sudo modprobe br_netfilter
    sudo apt-get -y update
    echo -e 'br_netfilter' | cat > /etc/modules-load.d/k8s.conf
    echo -e 'net.bridge.bridge-nf-call-ip6tables = 1\nnet.bridge.bridge-nf-call-iptables = 1' | cat >> /etc/sysctl.d/k8s.conf
    sysctl --system
}
#----------------------------------------------------------------------
docker_install(){
    echo '*****INSTALLING DOCKER*****'
    sudo apt-get update -y
    if  [ "$(which docker)" != "" ]
    then
        echo "Docker is already installed" 
    else
        sudo apt-get install docker.io -y 
    fi
	if [ "$(which vim)" != "" ]
    then
        echo "VIM is already installed" 
    else
        sudo apt-get install vim -y 
    fi
}	
#----------------------------------------------------------------------------
kube_packages(){
    echo '*****INSTALLING KUBEADM, KUBELET AND KUBECTL*****'
    sudo apt-get update -y
    if  [ "$(which apt-transport-https ca-certificates curl)" != "" ]
    then
        echo "apt-transport-https ca-certificates curl is already installed"  
    else
        sudo apt-get install apt-transport-https ca-certificates curl -y  
    fi
	sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
    echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
    sudo apt-get update
    apt-get install -y kubelet=$v kubectl=$v kubeadm=$v
    sudo apt-mark hold kubelet kubeadm kubectl
    systemctl start docker.service
    systemctl enable docker.service
}
#----------------------------------------------------------------------------
kube_cluster(){
    echo '*****STARTING CLUSTER USING KUBEADM*****'
    echo y | kubeadm reset  # To remove the existing cluster if running
    sudo rm -rf $HOME/.kube/config
    kubeadm init
    sleep 5s
    export KUBECONFIG=/etc/kubernetes/admin.conf
    sleep 5s
    systemctl restart kubelet
    sleep 20s
    export kubever=$(kubectl version | base64 | tr -d '\n')
    kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$kubever"
    sleep 20s
    mkdir -p $HOME/.kube
    sudo cp /etc/kubernetes/admin.conf $HOME/
    sudo chown $(id -u):$(id -g) $HOME/admin.conf
    export KUBECONFIG=/etc/kubernetes/admin.conf
    sleep 20s
}
#--------------------------------------------------------------------------------
golang_tools(){
    echo '*****INSTALLING GOLANG TOOLS FOR CLOUDCORE AND EDGECORE*****'
    sudo apt -y install make gcc jq
    wget https://dl.google.com/go/go1.14.15.linux-amd64.tar.gz -P /tmp
    tar -C /usr/local -xzf /tmp/go1.14.15.linux-amd64.tar.gz
    echo -e 'export PATH=$PATH:/usr/local/go/bin\nexport GOPATH=/usr/local/go/bin\nexport KUBECONFIG=/etc/kubernetes/admin.conf' |cat >> ~/.bashrc
    source $HOME/.bashrc
    sudo cp /usr/local/go/bin/go  /usr/local/bin
    kubectl get nodes
}
#---------------------------------------------------------------------------------
fornax_setup(){
    echo '*****FORNAX CONFIGURATION*****'
    mkdir -p $HOME/go/src/github.com/
    pushd $HOME/go/src/github.com/
    if [ "$(ls $HOME/go/src/github.com/)" == "" ]
    then
      git clone https://github.com/CentaurusInfra/fornax.git
    else
      sudo rm -rf fornax && git clone https://github.com/CentaurusInfra/fornax.git
    fi
    pushd $HOME/go/src/github.com/fornax
    systemctl restart docker
    cp $HOME/adminfile/admin.conf $HOME/go/src/github.com/fornax
    mkdir -p /etc/kubeedge/
    chmod a+x Makefile
    make all
    make WHAT=edgecore
    mkdir /etc/kubeedge/config -p
    echo 'SETTING UP THE EDGECORE'
    sudo cp /etc/kubernetes/admin.conf $HOME/edgecluster.kubeconfig
    _output/local/bin/edgecore --edgeclusterconfig > /etc/kubeedge/config/edgecore.yaml
    sed -i 's+RANDFILE+#RANDFILE+g' /etc/ssl/openssl.cnf
    tests/edgecluster/hack/update_edgecore_config.sh admin.conf
    echo 'APPLYING DEVICES.YAML'
    kubectl apply -f build/crds/devices/devices_v1alpha2_device.yaml
    kubectl apply -f build/crds/devices/devices_v1alpha2_devicemodel.yaml
    kubectl apply -f build/crds/reliablesyncs/cluster_objectsync_v1alpha1.yaml
    kubectl apply -f build/crds/reliablesyncs/objectsync_v1alpha1.yaml
    kubectl apply -f  build/crds/router/router_v1_rule.yaml
    kubectl apply -f  build/crds/router/router_v1_ruleEndpoint.yaml
    kubectl apply -f build/crds/edgecluster/mission_v1.yaml
    kubectl apply -f build/crds/edgecluster/edgecluster_v1.yaml
    chmod 777 $HOME/go/src/github.com/fornax/_output/local/bin/kubectl/vanilla/kubectl
    export KUBECONFIG=/etc/kubernetes/admin.conf
    nohup _output/local/bin/edgecore --edgecluster >> edgecore.logs 2>&1 &
    export KUBECONFIG=/etc/kubernetes/admin.conf
}
#--------------------------------------------------------------------------------------
processes(){
   if `[ !-z "$edgecore"]`
   then
      echo edgecore process is not running 
   else
      kill -9 $edgecore
      echo edgecore process killed forcefully, process id $edgecore.
   fi
}
#-------------------------------------------------------------------------------------
processes
#----------------
ip_tables
#----------------
docker_install
#----------------
kube_packages
#-----------------
kube_cluster
#-----------------
if [ "$(go version)" != "go version go1.14.15 linux/amd64" ]
then
    golang_tools
else
   echo " go1.14.15 already installed "
fi
#----------------
fornax_setup
#------------------------------------------------------------------------------
echo '*****SETUP SUCCESSSFUL*****' 
echo 'Logs: '
echo 'Edgecore: $HOME/go/src/github.com/fornax/edgecore.logs'
