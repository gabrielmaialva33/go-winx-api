# Name of the binary to generate
BINARY_NAME=go-winx-api
PORT ?= 8080  # Default port
HOST ?= localhost  # Default host

# Directory where the built binary will be stored
OUTPUT_DIR=bin

# Build configuration flags
GO_FLAGS=-mod=readonly
GO_LDFLAGS=-w -s  # Optimize binary (strip debugging and reduce size)
GO_TEST_FLAGS=-v  # Verbose flag for test output

# Directories for tests (optional)
TEST_DIR=./...

.PHONY: all build run test lint clean

# Default target (if no specific target is provided)
all: build

# Build the project binary
build:
	@echo "Building the project binary..."
	@mkdir -p $(OUTPUT_DIR)  # Ensure the output directory exists
	@go build $(GO_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME) ./main.go
	@echo "Binary generated at $(OUTPUT_DIR)/$(BINARY_NAME)"

# Run the server (builds first, then runs)
run: build
	@echo "Starting the server..."
	./$(OUTPUT_DIR)/$(BINARY_NAME) --host $(HOST) --port $(PORT)

# Run all test cases
test:
	@echo "Running tests..."
	@go test $(GO_TEST_FLAGS) $(TEST_DIR)

# Lint and format code, tidy dependencies
lint:
	@echo "Linting and verifying project dependencies..."
	@gofmt -l -w .  # Format all code (write changes to files automatically)
	@go mod tidy  # Ensure the go.mod and go.sum are up-to-date
	@golangci-lint run || echo "Linter finished with warnings or errors"

# Clean up generated files and binaries
clean:
	@echo "Cleaning up project artifacts..."
	@rm -rf $(OUTPUT_DIR)  # Remove the output directory
	@echo "Cleanup complete!"