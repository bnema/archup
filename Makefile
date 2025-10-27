# ArchUp Go Installer Makefile
# Builds the Go-based installer with proper version injection

# Get version from git tag or use dev
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
# Append branch name if not on main
CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
ifeq ($(CURRENT_BRANCH),dev)
	VERSION := $(VERSION)-dev
endif
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
LDFLAGS_STRIP := -ldflags "-X main.version=$(VERSION) -s -w"

# Build settings
BINARY_NAME := archup-installer
BUILD_DIR := build
CMD_DIR := cmd/archup-installer
INSTALL_DIR := /usr/local/bin

# Go build flags
GOFLAGS := -trimpath
CGO_ENABLED := 0

.PHONY: all build clean install uninstall run test fmt vet lint mod-tidy help

all: build

## build: Build the installer binary with version injection
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-release: Build optimized release binary (stripped, smaller size)
build-release:
	@echo "Building release $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(GOFLAGS) $(LDFLAGS_STRIP) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

## build-static: Build fully static binary (no dependencies)
build-static:
	@echo "Building static $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(GOFLAGS) $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Static build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install the binary to system
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed successfully"

## uninstall: Remove the binary from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstalled successfully"

## run: Build and run the installer
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

## fmt: Format all Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## lint: Run golangci-lint (install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run ./...

## mod-tidy: Tidy go.mod
mod-tidy:
	@echo "Tidying go.mod..."
	@go mod tidy

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@echo "Clean complete"

## version: Show version information
version:
	@echo "Version: $(VERSION)"

## help: Show this help message
help:
	@echo "ArchUp Go Installer - Makefile Commands"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
