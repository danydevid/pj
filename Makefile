# pj — Makefile
#
# Targets: build, install, clean, test, fmt, vet, lint, run
#
# Overrides: PREFIX, BIN_DIR, PLUGIN_DIR

PREFIX     ?= $(HOME)/.local
BIN_DIR    ?= $(PREFIX)/bin
PLUGIN_DIR ?= $(HOME)/.pj/bin

GO      ?= go
GOFLAGS ?=

.PHONY: all build install clean test fmt vet lint run help

all: build

## build: compile manager + plugins into ./bin
build:
	@scripts/build.sh

## install: install into $(BIN_DIR) and $(PLUGIN_DIR)
install: build
	@PREFIX=$(PREFIX) BIN_DIR=$(BIN_DIR) PLUGIN_DIR=$(PLUGIN_DIR) scripts/install.sh

## clean: remove build artifacts
clean:
	rm -rf bin

## test: run unit tests
test:
	$(GO) test $(GOFLAGS) ./...

## fmt: format all Go sources
fmt:
	@gofmt -s -w .
	@command -v goimports >/dev/null 2>&1 && goimports -w . || true

## vet: run go vet
vet:
	$(GO) vet ./...

## lint: run golangci-lint
lint:
	@golangci-lint run ./...

## run: build and invoke the manager (usage: make run ARGS="list")
run: build
	@./bin/pj $(ARGS)

## help: list targets
help:
	@awk 'BEGIN{FS=":.*##"} /^## /{sub(/^## /,""); print}' $(MAKEFILE_LIST)
	@echo
	@echo "Variables:"
	@echo "  PREFIX      = $(PREFIX)"
	@echo "  BIN_DIR     = $(BIN_DIR)"
	@echo "  PLUGIN_DIR  = $(PLUGIN_DIR)"