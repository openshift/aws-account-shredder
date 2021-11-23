#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

# unset the unbound variable error reporting cause we're explicityly checking the existence to provide a clear error message
set +u
if [ -z "$1" ]; then
    echo "Usage: $0 OUTPUT_FOLDER [GIT_REF]"
    exit 1
fi
set -u

yaml_out_folder="$1"
mkdir -p "${yaml_out_folder}"

#root_url="https://raw.githubusercontent.com/openshift/aws-account-operator/master"
root_url="https://raw.githubusercontent.com/mrWinston/aws-account-operator/osd-7582"
files_to_download=(
    deploy/crds/aws.managed.openshift.io_accounts.yaml
    deploy/service_account.yaml
    deploy/aas_role_binding.yaml
    deploy/aas_role.yaml
    deploy/aas_service_account.yaml
)

for file in ${files_to_download[@]}; do 
    printf "Downloading $(basename $file) to ${yaml_out_folder}\n"
    curl -Lo "${yaml_out_folder}/$(basename ${file})" --fail "${root_url}/${file}" &> /dev/null
done
