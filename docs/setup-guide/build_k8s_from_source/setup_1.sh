#!/bin/bash

set -e

echo -e "Setup Part 1 started... \n"

echo -e "Configuring iptables..."
sudo modprobe br_netfilter
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system
echo -e "[x] Finished configuring iptable.\n"

echo -e "Installing runtime docker..."
sudo apt-get -y update; sudo apt-get -y install docker.io
sudo systemctl enable docker.service
sudo usermod -aG docker $USER

echo -e "\nPlease reboot the machine."
echo -e "After reboot run Setup Part 2 script setup_2.sh"