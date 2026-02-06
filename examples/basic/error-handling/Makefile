# Error Handling - Makefile
# Category: Basic

.PHONY: run build test clean help

# Binary name (matches directory name)
BINARY := $(notdir $(CURDIR))

# Build directory
BUILD_DIR := bin

## run: Run the example directly
run:
	@echo "Running error-handling..."
	@go run main.go

## build: Build the example binary
build:
	@echo "Building error-handling..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY) main.go
	@echo "Binary created at $(BUILD_DIR)/$(BINARY)"

## test: Run tests (if any)
test:
	@echo "Running tests..."
	@go test -v .

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)/
	@echo "Clean complete"

## help: Show this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
