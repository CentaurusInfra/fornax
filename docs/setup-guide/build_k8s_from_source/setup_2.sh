#!/bin/bash

echo "Setup Part 2 started...\n"

echo "Continuing configuring docker..."
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
echo "[x] Finished setting up docker.\n"

echo "Starting installing Go..."
cd $HOME
sudo apt-get -y install make gcc jq
wget https://redirector.gvt1.com/edgedl/go/go1.17.1.linux-amd64.tar.gz -O go1.17.1.tar.gz
tar -xzvf go1.17.1.tar.gz
mkdir gopath
cat <<EOF | tee -a $HOME/.bashrc
export GOROOT=$HOME/go
export GOPATH=$HOME/gopath
export PATH=$HOME/go/bin:$PATH
EOF
. $HOME/.bashrc
echo "[x] Finished setting up Go."
echo "Rebooting the machine..."
echo "After reboot run Setup Part 3 script setup_3.sh..."
sleep 3
sudo reboot