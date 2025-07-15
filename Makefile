# p6s - PostgreSQL Terminal Management Tool
# Makefile for cross-platform builds

# Variables
APP_NAME = p6s
CMD_PATH = cmd/p6s/main.go
BUILD_DIR = build
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

# Default target
.PHONY: all
all: clean build-all

# Build for current platform
.PHONY: build
build:
	@echo "Building $(APP_NAME) for current platform..."
	go build $(LDFLAGS) -o $(APP_NAME) $(CMD_PATH)

# Build for all platforms
.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-windows-amd64 build-darwin-amd64 build-darwin-arm64
	@echo "All builds completed successfully!"

# Linux AMD64
.PHONY: build-linux-amd64
build-linux-amd64:
	@echo "Building for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_PATH)

# Linux ARM64
.PHONY: build-linux-arm64
build-linux-arm64:
	@echo "Building for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(CMD_PATH)

# Windows AMD64
.PHONY: build-windows-amd64
build-windows-amd64:
	@echo "Building for Windows AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)

# macOS Intel
.PHONY: build-darwin-amd64
build-darwin-amd64:
	@echo "Building for macOS Intel..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PATH)

# macOS Apple Silicon
.PHONY: build-darwin-arm64
build-darwin-arm64:
	@echo "Building for macOS Apple Silicon..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(APP_NAME)
	@rm -f $(APP_NAME)-*

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	go run $(CMD_PATH)

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build              - Build for current platform"
	@echo "  build-all          - Build for all supported platforms"
	@echo "  build-linux-amd64  - Build for Linux AMD64"
	@echo "  build-linux-arm64  - Build for Linux ARM64"
	@echo "  build-windows-amd64- Build for Windows AMD64"
	@echo "  build-darwin-amd64 - Build for macOS Intel"
	@echo "  build-darwin-arm64 - Build for macOS Apple Silicon"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Install dependencies"
	@echo "  test               - Run tests"
	@echo "  run                - Run the application"
	@echo "  help               - Show this help message"