#!/bin/bash

gateway_filename=cluster_gateway.properties

gateway_host_ip="$(cat ${gateway_filename})"

# Remove the old configmap file
rm -f cluster_gatewway_configmap.yaml

# Create a new configmap file from portal)host.properties
echo "Replacing [CLUSTER HOST] with $gateway_host_ip in configmap"
cat cluster_gateway_configmap.yaml.template|sed "s/CLUSTER_GATEWAY/$gateway_host_ip/" > cluster_gateway_configmap.yaml

# Use kubectl to apply changes
kubectl apply -f ./cluster_gateway_configmap.yaml

rm -f cluster_gateway_configmap.yaml

