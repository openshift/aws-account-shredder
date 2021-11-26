#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

repo_root="$(git rev-parse --show-toplevel)"

export AWS_ACCESS_KEY_ID_B64=$(printf "${AWS_ACCESS_KEY_ID}" | base64)
export AWS_SECRET_ACCESS_KEY_B64=$(printf "${AWS_SECRET_ACCESS_KEY}" | base64)

oc process --local \
    -f "${repo_root}/hack/templates/credentials-template.yaml" \
    -p "SHREDDER_NAMESPACE=${SHREDDER_NAMESPACE}" \
    -p "AWS_ACCESS_KEY_ID_B64=${AWS_ACCESS_KEY_ID_B64}" \
    -p "AWS_SECRET_ACCESS_KEY_B64=${AWS_SECRET_ACCESS_KEY_B64}" |
    oc apply -f -
