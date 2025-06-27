# Stampede Shooter Makefile

BINARY_NAME=stampede-shooter
GO_MODULE=stampede-shooter
BUILD_DIR=./build
CMD_DIR=./cmd/shooter

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "Built binaries for multiple platforms in $(BUILD_DIR)/"

# Install dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Run a quick smoke test
.PHONY: smoke-test
smoke-test: build
	@echo "Running smoke test..."
	@$(BUILD_DIR)/$(BINARY_NAME) --users 3 --script examples/smoke.yml --duration 5s --verbose

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)

# Install to system (requires sudo on Linux/macOS)
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed successfully"

# Analyze browser recording for header requirements
.PHONY: analyze-headers
analyze-headers:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make analyze-headers FILE=path/to/recording.json"; \
		echo "Example: make analyze-headers FILE=examples/sample-recording.json"; \
	else \
		echo "Analyzing browser recording: $(FILE)"; \
		go run tools/analyze-headers.go $(FILE); \
	fi

# Show help
.PHONY: help
help:
	@echo "Stampede Shooter - Load Testing Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build the binary"
	@echo "  build-all      Build for multiple platforms"
	@echo "  deps           Download and tidy dependencies"
	@echo "  test           Run tests"
	@echo "  smoke-test     Run a quick smoke test"
	@echo "  analyze-headers Analyze browser recording (requires FILE=path)"
	@echo "  clean          Clean build artifacts"
	@echo "  install        Install to /usr/local/bin (requires sudo)"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build"
	@echo "  make smoke-test"
	@echo "  make analyze-headers FILE=examples/sample-recording.json"
	@echo "  make build-all"

test-acme-demo: build
	@echo "Testing Acme demo with Rails authentication..."
	./build/stampede-shooter --users 2 --script examples/acme-demo.yml --duration 10s --verbose

test-acme-real: build
	@echo "Testing real Rails application (requires valid credentials)..."
	@echo "⚠️  Make sure to update examples/acme-rails-auth.yml with your application URL and examples/credentials.txt with your credentials first!"
	./build/stampede-shooter --users 3 --script examples/acme-rails-auth.yml --duration 30s --verbose

analyze-acme: tools/analyze-headers.go
	@echo "Analyzing Acme browser recording..."
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make analyze-acme FILE=your-recording.json"; \
		exit 1; \
	fi
	go run tools/analyze-headers.go $(FILE)

test-credentials: build
	@echo "Testing with credentials file..."
	./build/stampede-shooter --users 3 --script examples/acme-demo.yml --credentials examples/credentials.txt --duration 10s --verbose

test-acme-with-creds: build
	@echo "Testing Acme with credentials (requires valid credentials file)..."
	@echo "⚠️  Make sure to update examples/credentials.txt with your real credentials first!"
	./build/stampede-shooter --users 2 --script examples/acme-rails-auth.yml --credentials examples/credentials.txt --duration 30s --verbose 