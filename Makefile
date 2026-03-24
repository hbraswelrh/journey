# SPDX-License-Identifier: Apache-2.0

# Pac-Man Makefile
# Single entry point for all build, test, lint, format, and
# validation commands per the project constitution.

BINARY    := pacman
CMD_DIR   := ./cmd/pacman
BUILD_DIR := .

GO        := go
GOFLAGS   :=
LDFLAGS   :=

GOLANGCI  := golangci-lint
CUE       := cue

.PHONY: all build test lint fmt schema-check clean help \
       web-data web-build web-dev web-clean

## all: Build and lint (default target)
all: build lint

## build: Compile the Pac-Man binary
build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY) $(CMD_DIR)

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

## clean: Remove built binary, test caches, and web artifacts
clean: web-clean
	rm -f $(BUILD_DIR)/$(BINARY)
	$(GO) clean -testcache

## help: Print this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | \
		sed 's/^## //' | \
		column -t -s ':'
