.PHONY: build build-all clean

# Binary name
BINARY_NAME=ollama-cli
BUILD_DIR=build

# Version from git tag, default to dev
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "dev")

# Build for the current platform
build:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X github.com/masgari/ollama-cli/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for all supported platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/masgari/ollama-cli/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/masgari/ollama-cli/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/masgari/ollama-cli/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/masgari/ollama-cli/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/masgari/ollama-cli/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Clean build directory
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	@go test -v -count=1 -race -timeout=10m ./...

# Install the binary
install: build
	@echo "Installing binary to /usr/local/bin/$(BINARY_NAME)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete!" 