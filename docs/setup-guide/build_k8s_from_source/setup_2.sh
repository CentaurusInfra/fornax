#!/bin/bash

set -e

echo -e "Setup Part 2 started...\n"

echo -e "Continuing configuring docker..."
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
sudo systemctl daemon-reload
sudo systemctl restart docker
echo -e "[x] Finished setting up docker.\n"

echo -e "Starting installing Go..."
GO_VERSION="1.17.1"
cd $HOME
sudo apt-get -y install make gcc jq
wget https://redirector.gvt1.com/edgedl/go/go${GO_VERSION}.linux-amd64.tar.gz -O go${GO_VERSION}.tar.gz
tar -xzvf go${GO_VERSION}.tar.gz
mkdir gopath
cat <<EOF | tee -a $HOME/.bashrc
export GOROOT=$HOME/go
export GOPATH=$HOME/gopath
export PATH=$HOME/go/bin:$PATH
EOF
. $HOME/.bashrc
echo -e "[x] Finished setting up Go."
echo -e "\nPlease reboot the machine."
echo -e "After reboot run Setup Part 3 script setup_3.sh"