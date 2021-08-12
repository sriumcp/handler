IMG ?= handler:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Before pushing, `make` or `make all` will run through most of the build tests
all: fmt vet staticcheck test coverage

fmt: ## Run go fmt against code.
	go fmt ./...

vet:
	go vet ./...
	
staticcheck: sc
	$(STATICCHECK) ./...

test:
	go test ./... -coverprofile=coverage.out

coverage:
	@echo "test coverage: $(shell go tool cover -func coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}')"

show-coverage:
	go tool cover -html=coverage.out

docker-build:
	docker build . -t ${IMG}

docker-push:
	docker push ${IMG}

STATICCHECK = $(shell pwd)/bin/staticcheck
sc: ## Download staticcheck locally if necessary.
	$(call go-get-tool,$(STATICCHECK),honnef.co/go/tools/cmd/staticcheck@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef