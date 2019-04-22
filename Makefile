BINARY_HELPER := workshop
BINARY := piper
BINARY_DIR := bin
RELEASE_DIR := release

PLATFORMS := windows linux darwin
OS = $(word 1, $@)

VERSION ?= 1.0.0
ARCH ?= amd64

PKGS := $(shell go list ./... | grep -v /vendor)
GOPATH_BIN := $(GOPATH)/bin

# DEFAULT
.PHONY: default
default: build release

# BUILD
.PHONY: build
build: clean_build ensure test
	go build -race -o $(BINARY_DIR)/$(BINARY) ./cmd/$(BINARY)
	go build -race -o $(BINARY_DIR)/$(BINARY_HELPER) ./cmd/$(BINARY_HELPER)

# RELEASE
.PHONY: release
release: clean_release ensure test $(PLATFORMS)

# INSTALL
.PHONY: install
install: build
	cp $(BINARY_DIR)/$(BINARY) $(GOPATH_BIN)

# TEST
.PHONY: test
test: lint
	go test ${PKGS}

# DEP ENSURE
.PHONY: ensure
ensure: $(DEP)
	dep ensure

# LINT
.PHONY: lint
lint: $(GOLANGCI_LINT)
	golangci-lint run

# CLEAN-BUILD
.PHONY: clean_build
clean_build:
	rm -rf $(BINARY_DIR)

# CLEAN-RELEASE
.PHONY: clean_release
clean_release:
	rm -rf $(RELEASE_DIR)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p $(RELEASE_DIR)
	GOOS=$(OS) GOARCH=$(ARCH)
	go build -race -o $(RELEASE_DIR)/$(BINARY)-$(VERSION)-$(OS)-$(ARCH) ./cmd/$(BINARY)

# Check for golangci-lint tool dependency
GOLANGCI_LINT := $(GOPATH_BIN)/golangci-lint
$(GOLANGCI_LINT):
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Check for dep tool dependency
DEP := $(GOPATH_BIN)/dep
$(DEP):
	go get -u github.com/golang/dep/cmd/dep
