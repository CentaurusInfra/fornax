#! /bin/bash

set -e

# Enter the IP address of the new host 
echo -e "## Enter Private IP ADDRESS of Host Machine:"
read ip_machine
echo -e "\n"
echo -e "## Enter Absolute path of your key-pair of Host Machine:"
read  key_pair
echo -e "\n"


#To kill running process of cloudcore and edgecore
cloudcore=`ps -aef | grep _output/local/bin/cloudcore | grep -v sh| grep -v grep| awk '{print $2}'`
edgecore=`ps -aef | grep _output/local/bin/edgecore | grep -v sh| grep -v grep| awk '{print $2}'`
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
if [ -f "${DIR}/common.sh" ]; then
source "${DIR}/common.sh"
fi

key_gen(){
   if [ "$(ls /root/.ssh/id_rsa.pub)" != "/root/.ssh/id_rsa.pub" ] > /dev/null 2>&1
   then
       echo -e "## GENERATING KEY AND COPYING THE KEY TO HOST MACHINE"
       chmod 600 $key_pair
       < /dev/zero ssh-keygen -q -N ""
       cat ~/.ssh/id_rsa.pub | ssh -o StrictHostKeyChecking=no -i $key_pair root@$ip_machine "cat >> ~/.ssh/authorized_keys"
   else
       cat ~/.ssh/id_rsa.pub | ssh -o StrictHostKeyChecking=no -i $key_pair root@$ip_machine "cat >> ~/.ssh/authorized_keys"
   fi
}

cloud_edge_process(){
   if `[ !-z "$cloudcore"]`
   then
      echo cloudcore process is not running
   else
      kill -9 $cloudcore
      echo cloudcore process killed forcefully, process id $cloudcore.
   fi
   if `[ !-z "$edgecore"]`
   then
      echo edgecore process is not running
   else
      kill -9 $edgecore
      echo edgecore process killed forcefully, process id $edgecore.
   fi
}

fornax_add_edge(){
    echo  "## COPYING THE KUBECONFIG FILE TO HOST MACHINE"
    ssh -t root@$ip_machine "mkdir -p $HOME/machine_2_admin_file" > /dev/null 2>&1
    scp -r /etc/kubernetes/admin.conf  $ip_machine:$HOME/machine_2_admin_file
    export KUBECONFIG=/etc/kubernetes/admin.conf
    nohup _output/local/bin/edgecore --edgecluster >> edgecore.logs 2>&1 &
    export KUBECONFIG=/etc/kubernetes/admin.conf
    nohup _output/local/bin/cloudcore >> cloudcore.logs 2>&1 &
}

key_gen

cloud_edge_process

fornax_add_edge

echo -e "## SETUP SUCCESSSFUL\n"
echo -e "## Logs:"
echo -e "Cloudcore: $HOME/go/src/github.com/fornax/cloudcore.logs"
echo -e "Edgecore: $HOME/go/src/github.com/fornax/edgecore.logs\n"

