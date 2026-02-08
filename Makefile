.PHONY: build test clean run

# Binary name
BINARY_NAME=go-check
BINARY_DIR=bin

# Build the binary
build:
	mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/check

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	go clean
	rm -rf $(BINARY_DIR)

# Run the binary (example)
run:
	./$(BINARY_DIR)/$(BINARY_NAME) -threads=5 -duration=5

# Install dependencies
deps:
	go mod tidy

# Lint (requires golangci-lint)
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...