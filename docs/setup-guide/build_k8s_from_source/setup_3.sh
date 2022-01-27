#!/bin/bash

set -e

echo -e "Setup Part 3 started...\n"

echo -e "Starting building Kubernetes from source code..."
sudo apt -y install conntrack socat
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
git clone https://github.com/kubernetes/kubernetes.git
cd kubernetes
git fetch --all --tags
git checkout tags/v1.21.0
make clean
make all
CNI_VERSION="v0.8.2"
ARCH="amd64"
sudo mkdir -p /opt/cni/bin
curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz" | sudo tar -C /opt/cni/bin -xz
DOWNLOAD_DIR=/usr/local/bin
sudo mkdir -p $DOWNLOAD_DIR
CRICTL_VERSION="v1.22.0"
ARCH="amd64"
curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz" | sudo tar -C $DOWNLOAD_DIR -xz
for dir in /usr/bin/ /usr/local/bin/; do
    sudo ln -s $GOPATH/src/k8s.io/kubernetes/_output/bin/kubelet $dir
    sudo ln -s $GOPATH/src/k8s.io/kubernetes/_output/bin/kubeadm $dir
    sudo ln -s $GOPATH/src/k8s.io/kubernetes/_output/bin/kubectl $dir
done
RELEASE_VERSION="v0.4.0"
curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" | sed "s:/usr/bin:${DOWNLOAD_DIR}:g" | sudo tee /etc/systemd/system/kubelet.service
sudo mkdir -p /etc/systemd/system/kubelet.service.d
curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" | sed "s:/usr/bin:${DOWNLOAD_DIR}:g" | sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
sudo sed -i '5 i Environment="KUBELET_EXTRA_ARGS=--fail-swap-on=false"' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
sudo systemctl enable --now kubelet
echo -e "==============="
echo -e "Congratulations"
echo -e "==============="
echo -e "[x] Kubernetes development environment has been set up. Enjoy!"