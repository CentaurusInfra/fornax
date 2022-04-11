#!/bin/bash


slavehosts="./slaves.in"
cleanuplog=/tmp/cleanup.out
kubeadmresetlog=/tmp/kubeadm_reset.out
kubeadminitlog=/tmp/kubeadm_init.out
kubeadmjoinlog=/tmp/kubeadm_join.out
mizarlog=/tmp/mizar.out
secondstowait=10

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
if [ -f "${DIR}/common.sh" ]; then
source "${DIR}/common.sh"
fi


echo ">> verifying slave host file exist"
if [ ! -f "$slavehosts" ]; then
    echo "$slavehosts does not exist."
    echo "put IPs of slave hosts in a file called **slave.in** and try again."
    echo "exited on purpose."
    exit
fi

rm -f /tmp/*.out

nuke_slave(){
    slave=$1
    ssh -n $slave "kubeadm reset -f" > $kubeadmresetlog 2>&1 
    scp ./cleanup.sh $slave:/tmp > $cleanuplog 2>&1
    ssh -n $slave "/tmp/cleanup.sh" >> $cleanuplog 2>&1
}

echo ">> resetting and cleaning up slave nodes (in parallel)"
while IFS= read -r slave
do
    echo ">>>> nuking slave $slave"
    nuke_slave $slave & # remove & here to see log in its entirety
done < "$slavehosts"
wait

echo ">> resetting master node"
kubeadm reset -f | tee $kubeadmresetlog > $kubeadmresetlog 2>&1 

echo ">> starting master node"
kubeadm init --pod-network-cidr 20.0.0.0/16 > $kubeadminitlog 2>&1

echo ">> joining slave nodes (in parallel)"
token=`cat $kubeadminitlog|grep "\-\-token "|awk '{print $5}'`
master=`cat $kubeadminitlog|grep "\-\-token "|awk '{print $3}'`
certhash=`cat $kubeadminitlog|grep "discovery-token-ca-cert-hash"|awk '{print $2}'`
joincmd="kubeadm join $master --token $token --discovery-token-ca-cert-hash $certhash"

echo ">>>> joining with '$joincmd'"

while IFS= read -r slave
do
    echo ">>>> joining from $slave using $joincmd to $kubeadmjoinlog"
    ssh -n $slave "$joincmd" > $kubeadmjoinlog 2>&1 & # remove & here to see log in its entirety
    echo ">>> copying admin.conf to $slave"
    scp /etc/kubernetes/admin.conf $slave:/etc/kubernetes/
done < "$slavehosts"
wait 

echo ">> installing gateway configmap"
bash ./create_cluster_gateway_configmap.sh

echo ">> installing mizar"
kubectl create -f mizar.goose.yaml > $mizarlog 2>&1

echo ">> verifying vpc0 is provisioned"
vpc0=""
net0=""
while [[ "$vpc0" != *"Provisioned"* || "$net0" != *"Provisioned"* ]]
do
  echo ">>>> not yet. waiting $secondstowait seconds for vpc0 and its subnet0 to be provisioned"
  sleep $secondstowait 
  vpc0=`kubectl get vpc vpc0|awk '{print $6}'`
  net0=`kubectl get subnet net0|awk '{print $6}'`
done
echo "Update config files"
update_conf

echo "ALL DONE! YEEHAW!!"

echo ">>> sharing config with ubuntu"
cp /etc/kubernetes/admin.conf /home/ubuntu
chown ubuntu:ubuntu /home/ubuntu/admin.conf
