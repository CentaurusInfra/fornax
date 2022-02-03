#! /bin/bash

set -e

#Enter kubernetes version
v='1.21.1-00'

ip_tables(){
   echo '*****LETTING IPTABLES SEE BRIDGED TRAFFIC*****'
   sudo modprobe br_netfilter
   sudo apt-get -y update
   echo -e 'br_netfilter' | cat > /etc/modules-load.d/k8s.conf
   echo -e 'net.bridge.bridge-nf-call-ip6tables = 1\nnet.bridge.bridge-nf-call-iptables = 1' | cat >> /etc/sysctl.d/k8s.conf
   sysctl --system
}

docker_install(){
   echo '*****INSTALLING DOCKER*****'
   sudo apt-get update -y 
   if [ "$(which docker)" != "" ]
    then
       echo "Docker is already installed"
    else
       sudo apt-get install docker.io -y
   fi
}

kube_packages(){
   echo '*****INSTALLING KUBEADM, KUBELET AND KUBECTL*****'
   sudo apt-get update -y
   if [ "$(which apt-transport-https ca-certificates curl)" != "" ]
    then
      echo "apt-transport-https ca-certificates curl is already installed"
    else
      sudo apt-get install apt-transport-https ca-certificates curl -y
   fi
   sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
   echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
   sudo apt-get update 
   sudo apt-get install -y kubelet=$v kubectl=$v kubeadm=$v
   sudo apt-mark hold kubelet kubeadm kubectl 
   systemctl start docker.service
   systemctl enable docker.service
}

kube_cluster(){
   echo '*****STARTING CLUSTER USING KUBEADM*****'
   echo y | kubeadm reset # To remove the existing cluster if running
   sudo rm -rf $HOME/.kube/config
   kubeadm init
   systemctl restart kubelet
   export KUBECONFIG=/etc/kubernetes/admin.conf
   export kubever=$(kubectl version | base64 | tr -d '\n') 
   kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$kubever"
   mkdir -p $HOME/.kube
   sudo cp /etc/kubernetes/admin.conf $HOME/ 
   sudo chown $(id -u):$(id -g) $HOME/admin.conf
   kubectl get nodes
}

golang_tools(){
   if [ "$(go version)" != "go version go1.14.15 linux/amd64" ]
    then
       echo '*****INSTALLING GOLANG TOOLS FOR CLOUDCORE AND EDGECORE*****'
       sudo apt -y install make gcc jq
       wget https://dl.google.com/go/go1.14.15.linux-amd64.tar.gz -P /tmp
       tar -C /usr/local -xzf /tmp/go1.14.15.linux-amd64.tar.gz
       echo -e 'export PATH=$PATH:/usr/local/go/bin\nexport GOPATH=/usr/local/go/bin\nexport KUBECONFIG=/etc/kubernetes/admin.conf' |cat >> ~/.bashrc
       source $HOME/.bashrc
       sudo cp /usr/local/go/bin/go /usr/local/bin
    else
       echo " go1.14.15 already installed "
   fi
}

