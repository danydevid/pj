# Makefile
# Go Multi-Binary Project Automation Toolkit

# --- Configuration & Variables ---
APP_DIR    := ./cmd
BIN_DIR    := ./bin
APPS       := $(shell ls $(APP_DIR) 2>/dev/null)

# Go Commands
GOCMD      := go
GOBUILD    := $(GOCMD) build
GOTEST     := $(GOCMD) test
GOCLEAN    := $(GOCMD) clean

# Linter Tool
LINTER     := golangci-lint

# Build Flags
BUILD_FLAGS := -v

.PHONY: all help lint test build clean $(APPS)

# Default target executed when running 'make'
all: lint test build

## help: Display available CLI commands
help:
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  all       Run linter, test suite, and build all binaries"
	@echo "  lint      Execute static code analysis using .golangci.yml"
	@echo "  test      Run all unit tests in the workspace"
	@echo "  build     Compile all binaries in $(APP_DIR) to $(BIN_DIR)/"
	@echo "  clean     Remove compiled binaries and execution artifacts"
	@echo "  <app>     Build a specific application binary (e.g., make api)"

## lint: Run golangci-lint across all packages
lint:
	@echo "==> Running linters..."
	@$(LINTER) run ./...

## test: Run unit tests with race condition detector
test:
	@echo "==> Running unit tests..."
	@$(GOTEST) -v -race ./...

## build: Compile all binary targets in the cmd directory
build:
	@echo "==> Building all binaries..."
	@mkdir -p $(BIN_DIR)
	@for app in $(APPS); do \
		echo "  -> Compiling $$app..."; \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/$$app $(APP_DIR)/$$app; \
	done
	@echo "==> All binaries successfully created in $(BIN_DIR)/"

## <app>: Target for building an individual binary dynamically
$(APPS):
	@echo "==> Building dynamic target: $@..."
	@mkdir -p $(BIN_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/$@ $(APP_DIR)/$@
	@echo "==> Binary successfully created at $(BIN_DIR)/$@"

## clean: Remove all generated binaries and temporary files
clean:
	@echo "==> Cleaning build artifacts..."
	@$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@echo "==> Workspace cleaned successfully."