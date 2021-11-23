#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

repo_root="$(git rev-parse --show-toplevel)"

export AWS_ACCESS_KEY_ID_B64=$(printf "${AWS_ACCESS_KEY_ID}" | base64)
export AWS_SECRET_ACCESS_KEY_B64=$(printf "${AWS_SECRET_ACCESS_KEY}" | base64)

cat ${repo_root}/hack/templates/credentials.yaml.tpl | envsubst | oc apply -f -
