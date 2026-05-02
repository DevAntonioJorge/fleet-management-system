.PHONY: help generate build test clean

help:
	@echo "Available targets:"
	@echo "  generate     - Generate SQLc code from SQL queries"
	@echo "  build        - Build the project"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"

generate:
	@echo "Generating SQLc code..."
	@sqlc generate
	@echo "✓ SQLc code generated"

build: generate
	@echo "Building project..."
	@go build ./cmd/api
	@go build ./cmd/load-test
	@echo "✓ Build complete"

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning build artifacts..."
	@go clean
	@rm -f cmd/api/api cmd/load-test/load-test
	@echo "✓ Clean complete"
