#! /bin/bash

set -e

#Enter kubernetes version
v='1.21.1-00'
#Golang version
GO_VERSION=${GO_VERSION:-"1.13.9"}

ip_tables(){
   echo -e "## LETTING IPTABLES SEE BRIDGED TRAFFIC"
   sudo modprobe br_netfilter
   sudo apt-get -y update > /dev/null 2>&1
   echo -e 'br_netfilter' | cat > /etc/modules-load.d/k8s.conf
   echo -e 'net.bridge.bridge-nf-call-ip6tables = 1\nnet.bridge.bridge-nf-call-iptables = 1' | cat > /etc/sysctl.d/k8s.conf
   sysctl --system
   swapoff -a
   echo -e "## DONE\n"
}

docker_install(){
   sudo apt-get update -y > /dev/null 2>&1
   if [ "$(which docker)" != "" ] > /dev/null 2>&1
   then
     echo -e "## DOCKER IS ALREADY INSTALLED\n"
   else
     echo -e "##INSTALLING DOCKER"
     sudo apt -y install docker.io
     sudo mkdir -p /etc/docker
     cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
     systemctl enable docker
     systemctl daemon-reload
     systemctl restart docker
     echo -e "## DOCKER INSTALLED\n"
   fi
}

kube_packages(){
   echo -e "## INSTALLING KUBEADM, KUBELET AND KUBECTL"
   sudo apt-get update -y > /dev/null 2>&1
   if [ "$(which apt-transport-https ca-certificates curl)" != "" ] > /dev/null 2>&1
    then
      echo "apt-transport-https ca-certificates curl is already installed"
    else
      sudo apt-get install apt-transport-https ca-certificates curl -y
   fi
   sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
   echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
   sudo apt-get update > /dev/null 2>&1
   sudo apt-get install -y kubelet=$v kubectl=$v kubeadm=$v 
   sudo apt-mark hold kubelet kubeadm kubectl 
   systemctl start docker.service
   systemctl enable docker.service
   echo -e "## DONE\n"
}

kube_cluster(){
   if [ "$(ls /etc/kubernetes/admin.conf )" != "/etc/kubernetes/admin.conf" ] > /dev/null 2>&1
   then
      echo -e "## STARTING CLUSTER USING KUBEADM"
      kubeadm init
      systemctl restart kubelet
      export KUBECONFIG=/etc/kubernetes/admin.conf
      export kubever=$(kubectl version | base64 | tr -d '\n') 
      kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$kubever"
      sleep 6s
      kubectl get nodes
   else
      export KUBECONFIG=/etc/kubernetes/admin.conf
      kubectl get nodes
   fi
   echo -e "## KUBERNETES CLUSTER IS READY\n"
}

golang_tools(){
   if [ "$(go version)" != "go version go${GO_VERSION} linux/amd64" ] > /dev/null 2>&1
    then
       echo -e "## INSTALLING GOLANG TOOLS FOR CLOUDCORE AND EDGECORE"
       sudo apt -y install make gcc jq > /dev/null 2>&1
       wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz -P /tmp
       sudo tar -C /usr/local -xzf /tmp/go${GO_VERSION}.linux-amd64.tar.gz
       echo -e 'export PATH=$PATH:/usr/local/go/bin\nexport GOPATH=/usr/local/go/bin\nexport KUBECONFIG=/etc/kubernetes/admin.conf' |cat >> ~/.bashrc
       source $HOME/.bashrc
       sudo cp /usr/local/go/bin/go /usr/local/bin
       echo -e "## DONE\n"
    else
       echo -e "## go${GO_VERSION} already installed\n "
   fi
}

install_cni(){
  mkdir -p /etc/cni/net.d
  cat <<EOF | sudo tee /etc/cni/net.d/10-mizarcni.conf
{
  "cniVersion": "0.3.1",
  "name": "mizarcni",
  "type": "mizarcni"
}
EOF
}

update_conf() {
    cp /etc/kubernetes/admin.conf $HOME/.kube/config
    $HOME/go/src/github.com/fornax/_output/local/bin/cloudcore --minconfig > /etc/kubeedge/config/cloudcore.yaml
    cp /etc/kubernetes/admin.conf  $HOME/edgecluster.kubeconfig
    $HOME/go/src/github.com/fornax/_output/local/bin/edgecore --edgeclusterconfig > /etc/kubeedge/config/edgecore.yaml
    $HOME/go/src/github.com/fornax/tests/edgecluster/hack/update_edgecore_config.sh admin.conf
    kubectl apply -f $HOME/go/src/github.com/fornax/build/crds/devices/devices_v1alpha2_device.yaml
    kubectl apply -f $HOME/go/src/github.com/fornax/build/crds/devices/devices_v1alpha2_devicemodel.yaml
    kubectl apply -f $HOME/go/src/github.com/fornax/build/crds/reliablesyncs/cluster_objectsync_v1alpha1.yaml
    kubectl apply -f $HOME/go/src/github.com/fornax/build/crds/reliablesyncs/objectsync_v1alpha1.yaml
    kubectl apply -f $HOME/go/src/github.com/fornax/build/crds/router/router_v1_rule.yaml
    kubectl apply -f $HOME/go/src/github.com/fornax/build/crds/router/router_v1_ruleEndpoint.yaml
    kubectl apply -f $HOME/go/src/github.com/fornax/build/crds/edgecluster/mission_v1.yaml
    kubectl apply -f $HOME/go/src/github.com/fornax/build/crds/edgecluster/edgecluster_v1.yaml
    chmod 777 $HOME/go/src/github.com/fornax/_output/local/bin/kubectl/vanilla/kubectl
}
