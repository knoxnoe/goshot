.PHONY: all build clean test lint vendor ppa ppa-build ppa-sign ppa-upload

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
BINARY_NAME=goshot
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# PPA parameters
DISTRIBUTION=jammy
PPA_VERSION=0.6.1
PPA_ORIG_NAME=$(BINARY_NAME)_$(shell echo $(PPA_VERSION) | cut -d'-' -f1).orig.tar.gz
PPA_BUILD_DIR=/tmp/goshot-build
PPA_NAME=watzon/$(BINARY_NAME)

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/goshot

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf vendor/
	rm -rf $(PPA_BUILD_DIR)/

test:
	$(GOTEST) -v ./...

lint:
	$(GOVET) ./...

vendor:
	$(GOCMD) mod vendor

# PPA related targets
ppa: clean ppa-build ppa-sign ppa-upload
	@echo "Cleaning up build directory..."
	rm -rf $(PPA_BUILD_DIR)

ppa-build: vendor
	@echo "Building in $(PPA_BUILD_DIR)"
	# Create build directory
	rm -rf $(PPA_BUILD_DIR)
	mkdir -p $(PPA_BUILD_DIR)/$(BINARY_NAME)-$(PPA_VERSION)
	# Copy source files
	cp -a . $(PPA_BUILD_DIR)/$(BINARY_NAME)-$(PPA_VERSION)/
	cd $(PPA_BUILD_DIR)/$(BINARY_NAME)-$(PPA_VERSION) && rm -rf .git vendor
	# Create source tarball
	cd $(PPA_BUILD_DIR) && \
		tar -czf $(PPA_ORIG_NAME) $(BINARY_NAME)-$(PPA_VERSION)
	# Build the source package
	cd $(PPA_BUILD_DIR)/$(BINARY_NAME)-$(PPA_VERSION) && \
		debuild -S -sa -d --no-sign

ppa-sign:
	debsign "$(PPA_BUILD_DIR)/$(BINARY_NAME)_$(PPA_VERSION)-ubuntu.1_source.changes"

ppa-upload:
	dput --force ppa:$(PPA_NAME) "$(PPA_BUILD_DIR)/$(BINARY_NAME)_$(PPA_VERSION)-ubuntu.1_source.changes"

# Install build dependencies for debian packaging
ppa-deps:
	sudo apt-get install devscripts debhelper dput
