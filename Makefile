BUILD_SH := ./build/build.sh
.DEFAULT_GOAL := help

.PHONY: all help build vendor test-coverage lint install apis docker vet test clean

all: build ## Build all Go packages

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Build

build: ## Build all Go packages
	$(BUILD_SH) build

vendor: ## Vendor Go module dependencies
	$(BUILD_SH) vendor

##@ Test & Lint

test-coverage: ## Run tests, benchmarks, and coverage
	$(BUILD_SH) test-coverage

lint: ## Run golint checks
	$(BUILD_SH) lint

vet: ## Run go vet on all packages
	$(BUILD_SH) vet

test: ## Run vet and verbose unit tests
	$(BUILD_SH) test

##@ API

install: ## Install protoc-related generators
	$(BUILD_SH) install

apis: ## Generate protobuf/gateway/openapi assets
	$(BUILD_SH) apis

##@ Other

docker: ## Docker target placeholder
	$(BUILD_SH) docker

clean: ## Remove build artifacts
	$(BUILD_SH) clean

