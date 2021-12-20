#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

ctxName=$(oc config current-context)
config=$(oc config view -ojson)
clusterName=$(jq -r --arg CTX "$ctxName" '.contexts| .[] | select(.name==$CTX)| .context.cluster' <<< $config)
jq -r --arg CLUSTER "$clusterName" '.clusters | .[] | select(.name==$CLUSTER) | .cluster.server' <<< $config

