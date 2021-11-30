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

AAO_DEPENDENCIES_OUT_FOLDER=./hack/aao_dependencies

export NAMESPACE=aws-account-operator
export SHRED_ACCOUNT_NAME=aws-shredder-account-delete

# Ensure go modules are enabled:
export GOPROXY=https://proxy.golang.org

.PHONY: docker-build
docker-build: build


# Helper functions to improve local testing experience
.PHONY: create-account
create-account: check-shred-account-id
	cat hack/templates/aws-shredder-account-delete.yaml.tpl | envsubst | oc apply -f -

.PHONY: shred-account
shred-account:
	@osdctl account set $(SHRED_ACCOUNT_NAME) --state=Failed

.PHONY: delete-account
delete-account:
	@oc delete account -n $(NAMESPACE) $(SHRED_ACCOUNT_NAME)

.PHONY: get-logs
get-logs:
	oc -n $(NAMESPACE) logs -f deployment/aws-account-shredder

.PHONY: kill-pod
kill-pod:
	workflow/kill_pod.sh

.PHONY: check-shred-account-id
check-shred-account-id:
ifndef SHRED_ACCOUNT_ID
	$(error SHRED_ACCOUNT_ID is undefined)
endif

.PHONY: check-aws-account-credentials
check-aws-account-credentials: ## Check if AWS Account Env vars are set
ifndef AWS_ACCESS_KEY_ID
	$(error AWS_ACCESS_KEY_ID is undefined)
endif
ifndef AWS_SECRET_ACCESS_KEY
	$(error AWS_SECRET_ACCESS_KEY is undefined)
endif

.PHONY: download-aao-dependencies
download-aao-dependencies:
	hack/download_aao_dependencies.sh $(AAO_DEPENDENCIES_OUT_FOLDER)

.PHONY: apply-aao-dependencies
apply-aao-dependencies: download-aao-dependencies
	@for file in $$(find $(AAO_DEPENDENCIES_OUT_FOLDER) -type f -name '*.yaml'); do oc apply -f "$${file}"; done

.PHONY: create-shredder-credentials
create-shredder-credentials: check-aws-account-credentials
	hack/create_shredder_credentials.sh

.PHONY: create-namespace
create-namespace:
	@oc create ns $(NAMESPACE) || true

.PHONY: delete-namespace
delete-namespace:
	@oc delete ns $(NAMESPACE) || true

.PHONY: predeploy
predeploy: create-namespace apply-aao-dependencies create-shredder-credentials

.PHONY: redeploy
redeploy:delete-deploy deploy

.PHONY: deploy
deploy:
	oc apply -f ./deploy/deployment.yaml

.PHONY: delete-deploy
delete-deploy:
	oc delete -f ./deploy/deployment.yaml

.PHONY: clean-operator
clean-operator: delete-namespace
	rm -rf $(AAO_DEPENDENCIES_OUT_FOLDER)

GOLANGCI_LINT_CACHE ?= /tmp/golangci-cache

.PHONY: lint
lint: 
	GOLANGCI_LINT_CACHE=${GOLANGCI_LINT_CACHE}
	-@golangci-lint run 

.PHONY: test
test: 
	@go test ./...
