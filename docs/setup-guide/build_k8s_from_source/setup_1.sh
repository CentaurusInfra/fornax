#!/bin/bash

set -e

echo "Setup Part 1 started... \n"

echo "Configuring iptables..."
sudo modprobe br_netfilter
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system
echo "[x] Finished configuring iptable.\n"

echo "Installing runtime docker..."
sudo apt-get -y update; sudo apt-get -y install docker.io
sudo systemctl enable docker.service
sudo groupadd docker
sudo usermod -aG docker $USER

echo "\nRebooting the machine..."
echo "After reboot run Setup Part 2 script setup_2.sh"
sleep 3
sudo reboot