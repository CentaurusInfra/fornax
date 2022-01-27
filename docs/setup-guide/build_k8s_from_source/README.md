# Introduction
This guide introduces setting up a <b>development environment</b> for Kubernetes (K8s) cluster on a single instance (vm). `kubeadm` is installed for deploying and managing K8s clusters. This setup has been tested on a 2-node cluster.

# Launch a vm on AWS
- Click 'Launch Instances' on AWS
- Search 'ubuntu' and select 'Ubuntu Server 18.04 LTE (HVM), SSD Volume Type' with '64-bit (x86)'.
- Choose the Instance Type 't2.large', and click 'Next...'
- Change 'Auto-assign Public IP' to 'Enable' and click 'Next...'
- Chage storage size (GiB) to '100' and click 'Next...'
- Click 'Add Tag'; Type 'Name' for Key, and '\<vm-tag-name\>' for Value, check 'Instances' only. Click 'Next...'
- Click 'Add rule' for Inbound rules, following the below list
```
All TCP             TCP     0-65535     0.0.0.0/0
All UDP             UDP     0-65535     0.0.0.0/0
SSH                 TCP     22          0.0.0.0/0
All ICMP - IPv4     ICMP    All         0.0.0.0/0
```
- Click 'Review and Launch'
- Click 'Launch'
- Select a key pair
- Wait until the vm is running

# Install K8s components and tools
- SSH into the vm on AWS

## Configure iptables
- Run `sudo modprobe br_netfilter`
- Run `lsmod | grep br_netfilter` to make sure br_netfilter module is loaded
- Run
  ```
  cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
  br_netfilter
  EOF
  ```
- Run
  ```
  cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
  net.bridge.bridge-nf-call-ip6tables = 1
  net.bridge.bridge-nf-call-iptables = 1
  EOF
  ```
- Run `sudo sysctl --system`

## Install runtime docker
- Run `sudo apt-get -y update; sudo apt-get -y install docker.io`
- Run `sudo systemctl enable docker.service`
- Run `sudo groupadd docker`
- Run `sudo usermod -aG docker $USER`
- Reboot the machine by running `sudo reboot`
- Verify you can run `docker` without `sudo` by running `docker run hello-world`
- Run `sudo mkdir /etc/docker`
- Run 
```
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
```
- Run `sudo systemctl daemon-reload`
- Run `sudo systemctl restart docker`

## Install GoLang
- Run `GO_VERSION="1.17.1"`
- Run `cd ~`
- Run `sudo apt-get -y install make gcc jq`
- Run `wget https://redirector.gvt1.com/edgedl/go/go${GO_VERSION}.linux-amd64.tar.gz -O go${GO_VERSION}.tar.gz`
- Run `rm -rf /usr/local/go`
- Run `tar -xzvf go${GO_VERSION}.tar.gz`
- Run `mkdir gopath`
- Append the following lines to `~/.bashrc`
  ```
  export GOROOT=$HOME/go
  export GOPATH=$HOME/gopath
  export PATH=$GOROOT/bin:$PATH
  ```
- Run `source ~/.bashrc`
- Check `go version`

## Build Kubernetes from source
- Run `sudo apt -y install conntrack socat`
- Run `mkdir -p $GOPATH/src/k8s.io`
- Run `cd $GOPATH/src/k8s.io`
- Run `git clone https://github.com/kubernetes/kubernetes.git`
- Run `cd kubernetes`
- Run `git fetch --all --tags`
- Run `git checkout tags/v1.21.0`
- Run `make clean`
- Run `make all`
- Install CNI plugins by running:
```
CNI_VERSION="v0.8.2"
ARCH="amd64"
sudo mkdir -p /opt/cni/bin
curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz" | sudo tar -C /opt/cni/bin -xz
```
- Install crictl by running:
```
DOWNLOAD_DIR=/usr/local/bin
sudo mkdir -p $DOWNLOAD_DIR
CRICTL_VERSION="v1.22.0"
ARCH="amd64"
curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz" | sudo tar -C $DOWNLOAD_DIR -xz
```
- Run
  ```
  for dir in /usr/bin/ /usr/local/bin/; do
    sudo ln -s $GOPATH/src/k8s.io/kubernetes/_output/bin/kubelet $dir
    sudo ln -s $GOPATH/src/k8s.io/kubernetes/_output/bin/kubeadm $dir
    sudo ln -s $GOPATH/src/k8s.io/kubernetes/_output/bin/kubectl $dir
  done
  ```
- Add a `kubelet` systemd service by running
```
RELEASE_VERSION="v0.4.0"
curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" | sed "s:/usr/bin:${DOWNLOAD_DIR}:g" | sudo tee /etc/systemd/system/kubelet.service
sudo mkdir -p /etc/systemd/system/kubelet.service.d
curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" | sed "s:/usr/bin:${DOWNLOAD_DIR}:g" | sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
```
- Add `Environment="KUBELET_EXTRA_ARGS=--fail-swap-on=false"` in `/etc/systemd/system/kubelet.service.d/10-kubeadm.conf`
- Enable and start by running `sudo systemctl enable --now kubelet`
- Run
```
IPADDR=$(hostname -I | cut -d ' ' -f1)
NODENAME=$(hostname -s)
```
- Run `sudo kubeadm init --apiserver-advertise-address=$IPADDR  --apiserver-cert-extra-sans=$IPADDR  --pod-network-cidr=192.168.0.0/16 --node-name $NODENAME --ignore-preflight-errors Swap`
- Run `mkdir -p $HOME/.kube`
- Run `sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config`
- Run `sudo chown $(id -u):$(id -g) $HOME/.kube/config`
- `kubeadm` does not configure any network plugin, you need to install a network plugin of your choice. Here we are using Calico network plugin for this setup. Run `kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml`
- Verify by running `kubectl get po -n kube-system`
- Verify by running `kubectl get nodes`

# Automated Setup using Scripts (Optional)
- `sh setup_1.sh` and wait for rebooting
- `sh setup_2.sh` and wait for rebooting
- `sh setup_3.sh`
- After setup you need to run `kubeadm init` by yourself on the master node. You can repeat the same setup on worker nodes.

# Other references
- https://github.com/kubernetes/kubernetes/blob/master/hack/local-up-cluster.sh
- https://developer.ibm.com/articles/setup-guide-for-kubernetes-developers/
- https://github.com/kubernetes/kubernetes/issues/54918
