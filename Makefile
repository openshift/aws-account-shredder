GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)
GO=CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=on go
GOVERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1)
GOFLAGS ?=


# Ensure go modules are enabled:
export GOPROXY=https://proxy.golang.org

build:
	$(GO) build ${GOFLAGS} main.go