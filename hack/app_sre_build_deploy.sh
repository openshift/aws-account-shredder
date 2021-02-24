#!/bin/bash

# AppSRE team CD

set -exv

CURRENT_DIR=$(dirname "$0")

BASE_IMG="aws-account-shredder"
QUAY_IMAGE="quay.io/alima/${BASE_IMG}"
IMG="${QUAY_IMAGE}:latest"

GIT_HASH=$(git rev-parse --short=7 HEAD)

# build the image
BUILD_CMD="docker build" IMG="$IMG" make docker-build

docker push ${IMG}
# push the image
# skopeo copy --dest-creds "${QUAY_USER}:${QUAY_TOKEN}" \
#     "docker-daemon:${IMG}" \
#     "docker://${QUAY_IMAGE}:latest"

# skopeo copy --dest-creds "${QUAY_USER}:${QUAY_TOKEN}" \
#     "docker-daemon:${IMG}" \
#     "docker://${QUAY_IMAGE}:${GIT_HASH}"