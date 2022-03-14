#! /bin/bash

set -e

#To kill running process of cloudcore
cloudcore=`ps -aef | grep _output/local/bin/cloudcore | grep -v sh| grep -v grep| awk '{print $2}'`
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
if [ -f "${DIR}/common.sh" ]; then
source "${DIR}/common.sh"
fi

ip_tables

docker_install

kube_packages

golang_tools
