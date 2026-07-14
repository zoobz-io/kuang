.PHONY: build test test-unit lint lint-fix coverage clean help check ci install-tools install-hooks

.DEFAULT_GOAL := help

APP_NAME := kuang
BIN_DIR := bin

# Go modules in this workspace. Each is a separate module (own go.mod), so
# go test / golangci-lint must be run inside each one.
MODULES := . modules/github modules/matrix

help: ## Display available commands
	@echo "$(APP_NAME) Development Commands"
	@echo "================================"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Build
# =============================================================================
# kuang is a framework: `core.Run` is a library entrypoint a consumer's own
# main calls. The only binary in this repo is the MCP bridge (cmd/mcp).

build: ## Build the MCP bridge binary & the default Kuang server
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/mcp ./cmd/mcp
	@go build -o $(BIN_DIR)/kuang ./cmd/kuang

# =============================================================================
# Testing
# =============================================================================
# The `testing` build tag enables test-only hooks in dependencies (e.g.
# sctx.ResetAdminForTesting).

test: ## Run all tests with race detector (all modules)
	@for dir in $(MODULES); do \
		echo "==> test $$dir"; \
		(cd $$dir && go test -race -tags testing ./...) || exit 1; \
	done

test-unit: ## Run unit tests only (short mode, all modules)
	@for dir in $(MODULES); do \
		echo "==> test $$dir"; \
		(cd $$dir && go test -race -tags testing -short ./...) || exit 1; \
	done

# =============================================================================
# Code Quality
# =============================================================================

lint: ## Run linters (all modules)
	@for dir in $(MODULES); do \
		echo "==> lint $$dir"; \
		(cd $$dir && golangci-lint run --config=$(CURDIR)/.golangci.yml --timeout=5m) || exit 1; \
	done

lint-fix: ## Run linters with auto-fix (all modules)
	@for dir in $(MODULES); do \
		echo "==> lint-fix $$dir"; \
		(cd $$dir && golangci-lint run --config=$(CURDIR)/.golangci.yml --fix) || exit 1; \
	done

coverage: ## Generate coverage report (all modules)
	@for dir in $(MODULES); do \
		echo "==> coverage $$dir"; \
		(cd $$dir && go test -tags testing -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -1) || exit 1; \
	done

# =============================================================================
# Maintenance
# =============================================================================

clean: ## Remove generated files
	@rm -rf $(BIN_DIR) tmp
	@rm -f coverage.out coverage.html coverage.txt
	@rm -f modules/github/coverage.out modules/matrix/coverage.out
	@find . -name "*.test" -delete
	@find . -name "*.prof" -delete

install-tools: ## Install development tools
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2

install-hooks: ## Install git pre-commit hook
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make check' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed"

# =============================================================================
# CI
# =============================================================================

check: test lint ## Run tests and lint (quick validation)
	@echo "All checks passed!"

ci: clean lint test coverage ## Full CI simulation
	@echo "CI simulation complete!"
