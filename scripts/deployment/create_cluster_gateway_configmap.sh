#!/bin/bash

source cluster_gateway.properties

# Remove the old configmap file
rm -f cluster_gateway_configmap.yaml

# Create a new configmap file from portal)host.properties
echo "Replacing [CLUSTER HOST] with $gateway_name and $gateway_host_ip in configmap"
cat cluster_gateway_configmap.yaml.template|sed "s/GATEWAY_NAME/$gateway_name/" > cluster_gateway_configmap.yaml.tmp
cat cluster_gateway_configmap.yaml.tmp|sed "s/CLUSTER_GATEWAY/$gateway_host_ip/" > cluster_gateway_configmap.yaml


# Use kubectl to apply changes
kubectl apply -f ./cluster_gateway_configmap.yaml

rm -f cluster_gateway_configmap.yaml
rm -f cluster_gateway_configmap.yaml.tmp
