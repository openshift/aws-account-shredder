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


# Ensure go modules are enabled:
export GOPROXY=https://proxy.golang.org

.PHONY: docker-build
docker-build: build

