# Makefile for github.com/makegov/tango-go
#
# Common dev targets. Mirrors the script ergonomics of tango-node's
# package.json and tango-python's pyproject.toml.
#
# Note: the `integration` target points at ./tests/integration/, which does
# not exist in this wave. Phase 2 (tester) will create it. Running `make
# integration` before then is a no-op that prints "no Go files" — harmless.

GO       ?= go
PKG      ?= ./...
COVER    ?= coverage.out
COVERHTML?= coverage.html

.PHONY: all test test-race cover cover-html lint fmt vet tidy integration ci hooks help

# Default target: `make` runs the fast test suite.
all: test

test: ## Run unit tests
	$(GO) test $(PKG)

test-race: ## Run tests with the race detector
	$(GO) test -race $(PKG)

cover: ## Run tests with coverage; print per-function summary
	$(GO) test -coverprofile=$(COVER) $(PKG)
	$(GO) tool cover -func=$(COVER)

cover-html: cover ## Generate an HTML coverage report at coverage.html
	$(GO) tool cover -html=$(COVER) -o $(COVERHTML)
	@echo "open $(COVERHTML)"

lint: ## Run golangci-lint
	golangci-lint run $(PKG)

fmt: ## Format Go source in-place with gofmt -s
	gofmt -s -w .

vet: ## Run go vet
	$(GO) vet $(PKG)

tidy: ## go mod tidy
	$(GO) mod tidy

integration: ## Run integration tests (build tag `integration`)
	$(GO) test -tags=integration ./tests/integration/...

ci: vet test-race cover lint ## Everything CI runs locally

hooks: ## Install pre-commit + pre-push git hooks
	pre-commit install
	pre-commit install --hook-type pre-push

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
