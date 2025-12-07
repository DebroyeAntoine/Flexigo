# Flexigo Makefile
# Supports multiple OS and architectures

# Binaries
GO_BIN=bin/flexigo
RUST_BIN=bin/flexigo-tts

# Detect OS
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Default architecture
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Rust target mapping
ifeq ($(GOOS),linux)
    ifeq ($(GOARCH),amd64)
        RUST_TARGET=x86_64-unknown-linux-gnu
    else ifeq ($(GOARCH),arm64)
        RUST_TARGET=aarch64-unknown-linux-gnu
    else ifeq ($(GOARCH),arm)
        RUST_TARGET=armv7-unknown-linux-gnueabihf
    endif
    BIN_EXT=
    TTS_BIN_NAME=tts-rs
else ifeq ($(GOOS),darwin)
    ifeq ($(GOARCH),amd64)
        RUST_TARGET=x86_64-apple-darwin
    else ifeq ($(GOARCH),arm64)
        RUST_TARGET=aarch64-apple-darwin
    endif
    BIN_EXT=
    TTS_BIN_NAME=tts-rs
else ifeq ($(GOOS),windows)
    ifeq ($(GOARCH),amd64)
        RUST_TARGET=x86_64-pc-windows-gnu
    else ifeq ($(GOARCH),386)
        RUST_TARGET=i686-pc-windows-gnu
    endif
    BIN_EXT=.exe
    TTS_BIN_NAME=tts-rs.exe
endif

# Build flags
GO_BUILD_FLAGS=-ldflags="-s -w"
RUST_BUILD_FLAGS=--release

.PHONY: help
help: ## Show this help message
	@echo "Flexigo Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "Cross-compilation examples:"
	@echo "  make build-linux-amd64"
	@echo "  make build-linux-arm64"
	@echo "  make build-darwin-amd64"
	@echo "  make build-darwin-arm64"
	@echo "  make build-windows-amd64"
	@echo "  make build-all"

.PHONY: deps
deps: ## Install dependencies
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Checking Rust installation..."
	@which cargo > /dev/null || (echo "Rust not found. Install from https://rustup.rs" && exit 1)
	@echo "Dependencies ready!"

.PHONY: build-go
build-go: ## Build Go binary for current OS/arch
	@echo "Building Go binary for $(GOOS)/$(GOARCH)..."
	@mkdir -p bin
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GO_BUILD_FLAGS) -o $(GO_BIN)$(BIN_EXT) ./cmd/flexigo
	@echo "Go binary built: $(GO_BIN)$(BIN_EXT)"

.PHONY: build-rust
build-rust: ## Build Rust TTS binary for current OS/arch
	@echo "Building Rust TTS binary for $(RUST_TARGET)..."
	@mkdir -p bin
ifdef RUST_TARGET
	@rustup target add $(RUST_TARGET) 2>/dev/null || true
	cd rust/tts-rs && cargo build $(RUST_BUILD_FLAGS) --target=$(RUST_TARGET)
	cp rust/tts-rs/target/$(RUST_TARGET)/release/$(TTS_BIN_NAME) $(RUST_BIN)$(BIN_EXT)
else
	cd rust/tts-rs && cargo build $(RUST_BUILD_FLAGS)
	cp rust/tts-rs/target/release/$(TTS_BIN_NAME) $(RUST_BIN)$(BIN_EXT)
endif
	@echo "Rust binary built: $(RUST_BIN)$(BIN_EXT)"

.PHONY: build
build: build-go build-rust ## Build both Go and Rust binaries

.PHONY: run
run: build ## Build and run Flexigo
	@echo "Running Flexigo..."
	./$(GO_BIN)$(BIN_EXT)

.PHONY: test
test: ## Run all tests
	@echo "Running Go tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	cd rust/tts-rs && cargo clean

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting Rust code..."
	cd rust/tts-rs && cargo fmt

.PHONY: lint
lint: ## Run linters
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
	@echo "Running clippy..."
	cd rust/tts-rs && cargo clippy -- -D warnings

# Cross-compilation targets

.PHONY: build-linux-amd64
build-linux-amd64: ## Build for Linux AMD64
	@echo "Building for Linux AMD64..."
	@mkdir -p bin/linux-amd64
	GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o bin/linux-amd64/flexigo ./cmd/flexigo
	cd rust/tts-rs && cargo build $(RUST_BUILD_FLAGS) --target=x86_64-unknown-linux-gnu
	cp rust/tts-rs/target/x86_64-unknown-linux-gnu/release/tts-rs bin/linux-amd64/flexigo-tts
	@echo "Built: bin/linux-amd64/"

.PHONY: build-linux-arm64
build-linux-arm64: ## Build for Linux ARM64
	@echo "Building for Linux ARM64..."
	@mkdir -p bin/linux-arm64
	GOOS=linux GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o bin/linux-arm64/flexigo ./cmd/flexigo
	rustup target add aarch64-unknown-linux-gnu
	cd rust/tts-rs && cargo build $(RUST_BUILD_FLAGS) --target=aarch64-unknown-linux-gnu
	cp rust/tts-rs/target/aarch64-unknown-linux-gnu/release/tts-rs bin/linux-arm64/flexigo-tts
	@echo "Built: bin/linux-arm64/"

.PHONY: build-darwin-amd64
build-darwin-amd64: ## Build for macOS Intel
	@echo "Building for macOS Intel..."
	@mkdir -p bin/darwin-amd64
	GOOS=darwin GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o bin/darwin-amd64/flexigo ./cmd/flexigo
	rustup target add x86_64-apple-darwin
	cd rust/tts-rs && cargo build $(RUST_BUILD_FLAGS) --target=x86_64-apple-darwin
	cp rust/tts-rs/target/x86_64-apple-darwin/release/tts-rs bin/darwin-amd64/flexigo-tts
	@echo "Built: bin/darwin-amd64/"

.PHONY: build-darwin-arm64
build-darwin-arm64: ## Build for macOS Apple Silicon
	@echo "Building for macOS Apple Silicon..."
	@mkdir -p bin/darwin-arm64
	GOOS=darwin GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o bin/darwin-arm64/flexigo ./cmd/flexigo
	rustup target add aarch64-apple-darwin
	cd rust/tts-rs && cargo build $(RUST_BUILD_FLAGS) --target=aarch64-apple-darwin
	cp rust/tts-rs/target/aarch64-apple-darwin/release/tts-rs bin/darwin-arm64/flexigo-tts
	@echo "Built: bin/darwin-arm64/"

.PHONY: build-windows-amd64
build-windows-amd64: ## Build for Windows AMD64
	@echo "Building for Windows AMD64..."
	@mkdir -p bin/windows-amd64
	GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o bin/windows-amd64/flexigo.exe ./cmd/flexigo
	rustup target add x86_64-pc-windows-gnu
	cd rust/tts-rs && cargo build $(RUST_BUILD_FLAGS) --target=x86_64-pc-windows-gnu
	cp rust/tts-rs/target/x86_64-pc-windows-gnu/release/tts-rs.exe bin/windows-amd64/flexigo-tts.exe
	@echo "Built: bin/windows-amd64/"

.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 ## Build for all platforms
	@echo "All platforms built successfully!"
	@echo "Artifacts in bin/ directory:"
	@ls -lh bin/*/

.PHONY: install
install: build ## Install binaries to /usr/local/bin (requires sudo)
	@echo "Installing Flexigo..."
	sudo cp $(GO_BIN) /usr/local/bin/flexigo
	sudo cp $(RUST_BIN) /usr/local/bin/flexigo-tts
	@echo "Flexigo installed successfully!"
	@echo "Run with: flexigo"

.PHONY: uninstall
uninstall: ## Uninstall binaries from /usr/local/bin (requires sudo)
	@echo "Uninstalling Flexigo..."
	sudo rm -f /usr/local/bin/flexigo
	sudo rm -f /usr/local/bin/flexigo-tts
	@echo "Flexigo uninstalled successfully!"

.PHONY: release
release: clean build-all ## Prepare release artifacts
	@echo "Creating release archives..."
	@mkdir -p dist
	cd bin/linux-amd64 && tar -czf ../../dist/flexigo-linux-amd64.tar.gz flexigo flexigo-tts
	cd bin/linux-arm64 && tar -czf ../../dist/flexigo-linux-arm64.tar.gz flexigo flexigo-tts
	cd bin/darwin-amd64 && tar -czf ../../dist/flexigo-darwin-amd64.tar.gz flexigo flexigo-tts
	cd bin/darwin-arm64 && tar -czf ../../dist/flexigo-darwin-arm64.tar.gz flexigo flexigo-tts
	cd bin/windows-amd64 && zip ../../dist/flexigo-windows-amd64.zip flexigo.exe flexigo-tts.exe
	@echo "Release archives created in dist/"
	@ls -lh dist/
