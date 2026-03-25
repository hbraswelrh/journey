# SPDX-License-Identifier: Apache-2.0

# Gemara User Journey Makefile
# Single entry point for all build, test, lint, format, and
# validation commands per the project constitution.

GO        := go
GOFLAGS   :=

GOLANGCI  := golangci-lint
CUE       := cue

.PHONY: all test lint fmt schema-check clean help \
       web-data web-build web-dev web-clean

## all: Lint and build web interface (default target)
all: lint web-build

## test: Run all tests with race detector
test:
	$(GO) test -race -count=1 ./...

## lint: Run golangci-lint
lint:
	$(GOLANGCI) run ./...

## fmt: Format Go source files with goimports
fmt:
	goimports -w .

## schema-check: Validate output artifacts against Gemara CUE
##   schemas. Requires cue CLI and a local Gemara checkout.
schema-check:
	@echo "schema-check: no artifacts to validate yet"

## web-data: Generate TypeScript constants from Go source
web-data:
	$(GO) run ./cmd/genwebdata

## web-build: Build the web interface (runs codegen first)
web-build: web-data
	cd web && npm run build

## web-dev: Start the web dev server
web-dev: web-data
	cd web && npm run dev

## web-clean: Remove web build artifacts
web-clean:
	rm -rf web/dist web/src/generated

## clean: Remove test caches and web artifacts
clean: web-clean
	$(GO) clean -testcache

## help: Print this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | \
		sed 's/^## //' | \
		column -t -s ':'
