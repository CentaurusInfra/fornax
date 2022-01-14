#! /bin/bash




set -x

export b=
export c=

pushd /root   
hostnamectl set-hostname node-c
ufw disable
swapoff -a
#mkdir -p /etc/kubeedge/ca
#mkdir -p /etc/kubeedge/certs
apt-get -y update
echo -e 'br_netfilter' | cat > /etc/modules-load.d/k8s.conf
echo -e 'net.bridge.bridge-nf-call-ip6tables = 1\nnet.bridge.bridge-nf-call-iptables = 1' | cat >> /etc/sysctl.d/k8s.conf
sysctl --system
apt-get -y update
apt-get install docker.io -y
systemctl enable docker
systemctl start docker
apt-get -y update
apt-get install -y apt-transport-https ca-certificates curl
curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
echo -e 'deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main' | cat >> /etc/apt/sources.list.d/kubernetes.list
apt-get -y update
apt-get install -qy kubelet=1.21.1-00 kubectl=1.21.1-00 kubeadm=1.21.1-00
apt-mark hold kubelet kubeadm kubectl
systemctl enable docker.service
kubeadm init
export KUBECONFIG=/etc/kubernetes/admin.conf
sleep 120s
kubectl get nodes
export kubever=$(kubectl version | base64 | tr -d '\n')
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$kubever"
sleep 60s
kubectl get nodes
echo y | sudo apt-get install vim
GOLANG_VERSION=${GOLANG_VERSION:-"1.14.15"}
apt -y install make gcc jq
wget https://dl.google.com/go/go1.14.15.linux-amd64.tar.gz -P /tmp
tar -C /usr/local -xzf /tmp/go1.14.15.linux-amd64.tar.gz
echo -e 'export PATH=$PATH:/usr/local/go/bin\nexport GOPATH=/usr/local/go/bin\nexport KUBECONFIG=/etc/kubernetes/admin.conf' |cat >> ~/.bashrc 
sleep 10s
cp /usr/local/go/bin/go  /usr/local/bin
source /root/.bashrc
go version
mkdir -p go/src/github.com
pushd /root/go/src/github.com
git clone https://github.com/CentaurusInfra/fornax.git
         mv fornax kubeedge
         pushd /root/go/src/github.com/kubeedge
	 #ssh -t root@$b "echo yes | scp -r /etc/kubernetes/admin.conf  $c:/root/go/src/github.com/kubeedge"
	 #echo yes | scp -r $b:/etc/kubeedge/certs  /etc/kubeedge
         #echo yes | scp -r $b:/etc/kubeedge/ca  /etc/kubeedge
	 #echo yes | scp -r $b:/etc/kubernetes/admin.conf /root/go/src/github.com/kubeedge
	 cp /root/admin.conf /root/go/src/github.com/kubeedge
         make all
         make WHAT=edgecore
         mkdir /etc/kubeedge/config -p
         cp /etc/kubernetes/admin.conf /root/edgecluster.kubeconfig
         _output/local/bin/edgecore --edgeclusterconfig > /etc/kubeedge/config/edgecore.yaml
         tests/edgecluster/hack/update_edgecore_config.sh admin.conf
	 sed -i 's+RANDFILE+#RANDFILE+g' /etc/ssl/openssl.cnf
         kubectl apply -f build/crds/devices/devices_v1alpha2_device.yaml
         kubectl apply -f build/crds/devices/devices_v1alpha2_devicemodel.yaml
         kubectl apply -f build/crds/reliablesyncs/cluster_objectsync_v1alpha1.yaml
         kubectl apply -f build/crds/reliablesyncs/objectsync_v1alpha1.yaml
         kubectl apply -f  build/crds/router/router_v1_rule.yaml
         kubectl apply -f  build/crds/router/router_v1_ruleEndpoint.yaml
         kubectl apply -f build/crds/edgecluster/mission_v1.yaml
         kubectl apply -f build/crds/edgecluster/edgecluster_v1.yaml
         chmod 777 /root/go/src/github.com/kubeedge/_output/local/bin/kubectl/vanilla/kubectl
	 export KUBECONFIG=/etc/kubernetes/admin.conf
         nohup _output/local/bin/edgecore --edgecluster > edgecore.logs 2>&1 &
         sleep 5s
         cat edgecore.logs
		 
