# Makefile for EisenKan

SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eufo pipefail -c

curdir:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
rootdir:=$(shell git rev-parse --show-toplevel)

APP_NAME := eisenkan
BIN_DIR := bin
OUTPUT := $(BIN_DIR)/$(APP_NAME)
SRC_DIR := ./cmd/$(APP_NAME)

# Version can be set via environment variable: make build VERSION=1.0.0
VERSION ?= dev

# Build flags to suppress duplicate library warnings on macOS
ifeq ($(shell uname),Darwin)
	BUILD_FLAGS := -ldflags "-X main.version=$(VERSION) -w"
	export CGO_LDFLAGS := $(CGO_LDFLAGS) -Wl,-no_warn_duplicate_libraries
else
	BUILD_FLAGS := -ldflags "-X main.version=$(VERSION)"
endif

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(OUTPUT) $(SRC_DIR)

.PHONY: run
run: build ## Build and run the application
	@echo "Running $(APP_NAME)..."
	@$(OUTPUT)

.PHONY: test
test: build ## Run fast unit tests only
	@echo "Running unit tests for $(APP_NAME)..."
	go test -short -run "TestUnit_" ./...

.PHONY: test-integration
test-integration: build ## Run unit and integration tests (medium speed)
	@echo "Running unit and integration tests for $(APP_NAME)..."
	go test -short -run "Test(Unit_|Integration_)" ./...

.PHONY: test-acceptance
test-acceptance: build ## Run all tests including slow acceptance tests
	@echo "Running all tests for $(APP_NAME) (this may take several minutes)..."
	go test -run "Test(Unit_|Integration_|Acceptance_)" ./...

.PHONY: test-all
test-all: test-acceptance ## Run all tests (same as test-acceptance)

.PHONY: install
install: ## Install the application
	@echo "Installing $(APP_NAME) to $$GOBIN..."
	go install $(SRC_DIR)

.PHONY: clean
clean: ## Clean the build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
