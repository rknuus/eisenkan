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
SYNC_REPO := https://github.com/rknuus/idesign_project_template_sync.git
SYNC_PREFIX := 3rd-party/sync
SYNC_BRANCH := main

# Template repository configuration
TEMPLATE_USER := rknuus
TEMPLATE_REPO := idesign_project_template
TEMPLATE_BRANCH := main

# Version can be set via environment variable: make build VERSION=1.0.0
VERSION ?= dev

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT) $(SRC_DIR)

.PHONY: run
run: build ## Build and run the application
	@echo "Running $(APP_NAME)..."
	@$(OUTPUT)

# FIXME(RAKN): untested
test: build ## Run tests
	@echo "Running tests for $(APP_NAME)..."
	go test ./...

.PHONY: sync-setup
sync-setup: ## Setup the git subtrees
	git subtree add --prefix=$(SYNC_PREFIX) $(SYNC_REPO) $(SYNC_BRANCH) --squash

.PHONY: sync-update
sync-update: ## Synchronize the git subtrees
	git subtree pull --prefix=$(SYNC_PREFIX) $(SYNC_REPO) $(SYNC_BRANCH) --squash

.PHONY: sync-reset
sync-reset: ## Reset subtree to remote state (destructive)
	@echo "Resetting subtree to remote state..."
	rm -rf $(SYNC_PREFIX)
	git subtree add --prefix=$(SYNC_PREFIX) $(SYNC_REPO) $(SYNC_BRANCH) --squash
	@echo "Subtree reset complete"

.PHONY: sync-claude
sync-claude: sync-update ## Update CLAUDE.md from template repository
	@echo "Updating CLAUDE.md from template repository..."
	./$(SYNC_PREFIX)/scripts/update-claude-md.sh $(TEMPLATE_USER) $(TEMPLATE_REPO) $(TEMPLATE_BRANCH)

.PHONY: install
install: ## Install the application
	@echo "Installing $(APP_NAME) to $$GOBIN..."
	go install $(SRC_DIR)

.PHONY: clean
clean: ## Clean the build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
