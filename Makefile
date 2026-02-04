# ============================================
# Flexigo – Makefile Robuste (Go + Rust + NPM)
# ============================================

APP_NAME=flexigo
APP_ID=com.flexigo.app
APP_VERSION=1.0.0
DIST=release
BIN=bin

RUST_DIR=rust/tts-rs
BROWSER_DIR=browser
GO_MODULE=github.com/DebroyeAntoine/flexigo

BROWSER_BUILD_OUT=$(BIN)/browser

GREEN  := $(shell tput setaf 2)
YELLOW := $(shell tput setaf 3)
RESET  := $(shell tput sgr0)

.PHONY: help deps build-all clean

help:
	@echo "$(GREEN)make build-all$(RESET) - Compile pour Windows, Linux et Mac"

deps:
	go mod tidy
	go mod download
	rustup target add x86_64-pc-windows-gnu x86_64-apple-darwin aarch64-apple-darwin
	cd $(BROWSER_DIR) && npm install
	go install fyne.io/fyne/v2/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest

# --------------------------------------------
# WINDOWS
# --------------------------------------------
cross-windows:
	@echo "$(YELLOW)➜ Build Windows...$(RESET)"
	# Utilisation du chemin relatif au module pour éviter les erreurs de main undeclared
	fyne-cross windows --arch=amd64 --app-id=$(APP_ID) --app-version=$(APP_VERSION) --output=$(APP_NAME).exe ./cmd
	cd $(RUST_DIR) && cargo build --release --target=x86_64-pc-windows-gnu
	cd $(BROWSER_DIR) && npm run build:win -- --x64
	
	@mkdir -p $(DIST)/windows/browser
	cp fyne-cross/bin/windows-amd64/$(APP_NAME).exe $(DIST)/windows/
	cp $(RUST_DIR)/target/x86_64-pc-windows-gnu/release/tts-rs.exe $(DIST)/windows/flexigo-tts.exe
	cp -r $(BROWSER_BUILD_OUT)/win*-unpacked/* $(DIST)/windows/browser/

# --------------------------------------------
# LINUX
# --------------------------------------------
cross-linux:
	@echo "$(YELLOW)➜ Build Linux...$(RESET)"
	fyne-cross linux --arch=amd64 --app-id=$(APP_ID) --output=$(APP_NAME) --no-cache ./cmd
	
	@echo "$(YELLOW)➜ Build Rust Linux (Docker)...$(RESET)"
	docker run --rm --platform linux/amd64 -v $(shell pwd):/app -w /app/$(RUST_DIR) rust:1.82-bookworm /bin/bash -c "\
		apt-get update && apt-get install -y libspeechd-dev pkg-config clang libclang-dev && \
		cargo build --release"
	
	cd $(BROWSER_DIR) && npm run build:linux -- --x64
	
	@mkdir -p $(DIST)/linux/browser
	cp fyne-cross/bin/linux-amd64/$(APP_NAME) $(DIST)/linux/
	cp $(RUST_DIR)/target/release/tts-rs $(DIST)/linux/flexigo-tts
	cp -r $(BROWSER_BUILD_OUT)/linux*-unpacked/* $(DIST)/linux/browser/

# --------------------------------------------
# MAC
# --------------------------------------------
cross-mac:
	@echo "$(YELLOW)➜ Build macOS...$(RESET)"
	fyne-cross darwin --arch=amd64,arm64 --app-id=$(APP_ID) --output=$(APP_NAME) ./cmd
	cd $(RUST_DIR) && cargo build --release --target=x86_64-apple-darwin
	cd $(RUST_DIR) && cargo build --release --target=aarch64-apple-darwin
	cd $(BROWSER_DIR) && npm run build:mac
	
	@mkdir -p $(DIST)/macos
	cp -r fyne-cross/dist/darwin-arm64/$(APP_NAME).app $(DIST)/macos/
	lipo -create -output $(DIST)/macos/$(APP_NAME).app/Contents/Resources/flexigo-tts \
		$(RUST_DIR)/target/x86_64-apple-darwin/release/tts-rs \
		$(RUST_DIR)/target/aarch64-apple-darwin/release/tts-rs
	
	@mkdir -p $(DIST)/macos/$(APP_NAME).app/Contents/Resources/browser
	cp -r $(BROWSER_BUILD_OUT)/mac*/$(APP_NAME)-browser.app $(DIST)/macos/$(APP_NAME).app/Contents/Resources/browser/

build-all: clean cross-windows cross-linux cross-mac

clean:
	rm -rf $(DIST) fyne-cross/ $(BIN)/browser
	cd $(RUST_DIR) && cargo clean
