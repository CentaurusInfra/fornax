#!/bin/bash

# THis is a tool to update /etc/kubeedge/config/edgecore.yaml to work with a cloudcore whose kubeconfig file is specified in the command line

set -e

KUBECONFIG_FILE=$1

if [ -z "${KUBECONFIG_FILE}" ]; then 
    echo "Please specify the kubeconfig file to access the cloud cluster"
    exit
fi

if [ ! -f "${KUBECONFIG_FILE}" ]; then
    echo "kubeconfig file (${KUBECONFIG_FILE}) not found!"
    exit
fi

TARGET_FILE="/etc/kubeedge/config/edgecore.yaml"

IP_ADDRESS=$(grep server "${KUBECONFIG_FILE}" | awk -F "/" '{print $3}' | awk -F ":" '{print $1}')

TOKEN=$(kubectl get secret -nkubeedge tokensecret --kubeconfig "${KUBECONFIG_FILE}" -o=jsonpath='{.data.tokendata}' | base64 -d)

TOKEN_LINE="    token: \"${TOKEN}\""

sed -i "s/    token:.*/${TOKEN_LINE}/" ${TARGET_FILE}

SERVER_LINE="      server: ${IP_ADDRESS}:10000"

sed -i "s/      server:.*/${SERVER_LINE}/" ${TARGET_FILE}

HTTPSERVER_LINE="    httpServer: ${IP_ADDRESS}:10002"

sed -i "s/    httpServer:.*/${HTTPSERVER_LINE}/" ${TARGET_FILE}