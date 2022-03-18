SHELL := /usr/bin/env bash

OPERATOR_DOCKERFILE = ./deploy/Dockerfile
REUSE_UUID := $(shell uuidgen | awk -F- '{ print tolower($$2) }')
REUSE_BUCKET_NAME=test-reuse-bucket-${REUSE_UUID}


# Include shared Makefiles
include project.mk
include standard.mk


GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)
GO=CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=on go
GOVERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1)
GOFLAGS ?=

ACCOUNT_CRD_REF=master

export SHREDDER_NAMESPACE=aws-account-shredder
ACCOUNT_OPERATOR_NAMESPACE=aws-account-operator

IMAGE=$(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)/$(IMAGE_NAME)
IMAGE_TAG=latest

.PHONY: check-shred-account-id-file
check-shred-account-id-file:
ifndef AWS_ACCOUNTS_TO_SHRED_FILE
	$(error AWS_ACCOUNTS_TO_SHRED_FILE is undefined)
endif

.PHONY: check-aws-account-credentials
check-aws-account-credentials: ## Check if AWS Account Env vars are set
ifndef AWS_ACCESS_KEY_ID
	$(error AWS_ACCESS_KEY_ID is undefined)
endif
ifndef AWS_SECRET_ACCESS_KEY
	$(error AWS_SECRET_ACCESS_KEY is undefined)
endif

## Check, if we are using a kind or crc oc context to avoid making changes to prod envs
only-local-ctx:
	hack/get_current_api_url.sh | grep '127.0.0.1\|api.crc.testing'

# Ensure go modules are enabled:
export GOPROXY=https://proxy.golang.org

.PHONY: docker-build
docker-build: build

# Helper functions to improve local testing experience
.PHONY: shred-accounts
shred-accounts: only-local-ctx check-shred-account-id-file
	hack/shred_accounts.sh -f $(AWS_ACCOUNTS_TO_SHRED_FILE) mark

.PHONY: shred-accounts-status
shred-accounts-status: only-local-ctx check-shred-account-id-file
	hack/shred_accounts.sh -f $(AWS_ACCOUNTS_TO_SHRED_FILE) status

.PHONY: shred-accounts-cleanup
shred-accounts-cleanup: only-local-ctx check-shred-account-id-file
	hack/shred_accounts.sh -f $(AWS_ACCOUNTS_TO_SHRED_FILE) cleanup

.PHONY: get-logs
get-logs: only-local-ctx
	oc logs --follow deployment/aws-account-shredder -n $(SHREDDER_NAMESPACE)

.PHONY: kill-pod
kill-pod: only-local-ctx
	workflow/kill_pod.sh

GOLANGCI_LINT_CACHE ?= /tmp/golangci-cache

.PHONY: lint
lint:
	GOLANGCI_LINT_CACHE=${GOLANGCI_LINT_CACHE}
	-@golangci-lint run

.PHONY: test
test:
	@go test ./...

.PHONY: create-account-crd
create-account-crd: only-local-ctx
	curl https://raw.githubusercontent.com/openshift/aws-account-operator/$(ACCOUNT_CRD_REF)/deploy/crds/aws.managed.openshift.io_accounts.yaml | oc apply -f -

.PHONY: delete-account-crd
delete-account-crd: only-local-ctx
	curl https://raw.githubusercontent.com/openshift/aws-account-operator/$(ACCOUNT_CRD_REF)/deploy/crds/aws.managed.openshift.io_accounts.yaml | oc delete -f - || true

.PHONY: create-shredder-credentials
create-shredder-credentials: only-local-ctx check-aws-account-credentials
	hack/create_shredder_credentials.sh

.PHONY: create-namespace
create-namespace: only-local-ctx
	@oc create ns $(SHREDDER_NAMESPACE) || true
	@oc create ns $(ACCOUNT_OPERATOR_NAMESPACE) || true

.PHONY: delete-namespace
delete-namespace: only-local-ctx
	@oc delete ns $(SHREDDER_NAMESPACE) || true
	@oc delete ns $(ACCOUNT_OPERATOR_NAMESPACE) || true

.PHONY: predeploy
predeploy: only-local-ctx create-namespace create-account-crd create-shredder-credentials

.PHONY: redeploy
redeploy: only-local-ctx delete-deploy deploy

.PHONY: deploy
deploy: only-local-ctx
	@oc process --local \
		-f ./deploy/aws-account-shredder-template.yaml \
		-p "IMAGE=$(IMAGE)" \
		-p "IMAGE_TAG=$(IMAGE_TAG)" \
		-p "REPLICAS=1" | \
		oc apply -n "$(SHREDDER_NAMESPACE)" -f -
	@oc process --local \
		-f ./deploy/aws-account-operator-template.yaml \
		-p "SHREDDER_NAMESPACE=$(SHREDDER_NAMESPACE)" | \
		oc apply -n "$(ACCOUNT_OPERATOR_NAMESPACE)" -f -

.PHONY: delete-deploy
delete-deploy: only-local-ctx
	@oc process --local \
		-f ./deploy/aws-account-shredder-template.yaml \
		-p "IMAGE=$(IMAGE)" \
		-p "IMAGE_TAG=$(IMAGE_TAG)" \
		-p "REPLICAS=1" | \
		oc delete -n "$(SHREDDER_NAMESPACE)" -f -
	@oc process --local \
		-f ./deploy/aws-account-operator-template.yaml \
		-p "SHREDDER_NAMESPACE=$(SHREDDER_NAMESPACE)" | \
		oc delete -n "$(ACCOUNT_OPERATOR_NAMESPACE)" -f -

.PHONY: delete-operator
clean-operator: delete-namespace delete-account-crd
