#!/bin/bash

# THis is a tool to automatically update the token in /etc/kubeedge/config/edgecore.yaml

set -e

TARGET_FILE="/etc/kubeedge/config/edgecore.yaml"

TOKEN=$(kubectl get secret -nkubeedge tokensecret -o=jsonpath='{.data.tokendata}' | base64 -d)

TOKEN_LINE="    token: \"${TOKEN}\""

sed -i "s/    token:.*/${TOKEN_LINE}/" ${TARGET_FILE}
