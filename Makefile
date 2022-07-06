# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=/usr/bin/env bash

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

BIN_DIR := bin
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

GOPROXY=devops-centos7-atom1.cisco.com/artifactory/proxy-go-remote
GOPRIVATE=aci-github.cisco.com
GOBUILD=GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) go build

GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
GOIMPORTS := $(TOOLS_BIN_DIR)/goimports
GINKGO := $(TOOLS_BIN_DIR)/ginkgo

TAG ?= dev
ARCH ?= amd64
REGISTRY ?= infra
IMAGE_NAME ?= webex-bot
export WEBEX_BOT_IMG ?= $(REGISTRY)/$(IMAGE_NAME)

###############################################################################
## Tooling Binaries
###############################################################################

$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod # Build golangci-lint from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) github.com/golangci/golangci-lint/cmd/golangci-lint

$(GOIMPORTS):
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) golang.org/x/tools/cmd/goimports

$(GINKGO): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GOBUILD) -tags tools -o $(subst hack/tools/,,$@) github.com/onsi/ginkgo/v2/ginkgo	
	
.PHONY: tools
tools:  $(GOLANGCI_LINT) $(GOIMPORTS) $(GINKGO) ## build all tools

###############################################################################
## Help
###############################################################################

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

###############################################################################
## Linting
###############################################################################

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint codebase
	$(GOLANGCI_LINT) run -v --fast=false

###############################################################################
# Building the binary
###############################################################################

.PHONY: fmt
fmt goimports: $(GOIMPORTS) ## Format and adjust import modules.
	$(GOIMPORTS) -local github.com/gianlucam76 -w .

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

###############################################################################
# Tests
###############################################################################

.PHONY: ut
ut: $(GINKGO) ## Run uts
	$(GINKGO) --v --trace --randomize-all -r --label-filter=${FOCUS} 
