# ============================================
# Flexigo – Makefile Robuste (Go + Rust + NPM)
# ============================================

BIN=bin
RUST_DIR=rust/tts-rs
BROWSER_DIR=browser
# On pointe vers le dossier du package, pas le fichier
GO_PKG=./cmd
APP_NAME=flexigo
APP_ID=com.flexigo.app
APP_VERSION=1.0.0

# Couleurs
GREEN  := $(shell tput setaf 2)
YELLOW := $(shell tput setaf 3)
RESET  := $(shell tput sgr0)

.PHONY: help deps build-all clean

help:
	@echo "$(GREEN)Cibles :$(RESET) make deps, make build-all, make clean"

# --------------------------------------------
# 1. Dépendances & Préparation
# --------------------------------------------
deps:
	@echo "$(YELLOW)➜ Préparation des modules Go...$(RESET)"
	go mod tidy
	go mod download
	cd $(BROWSER_DIR) && npm install
	go install fyne.io/fyne/v2/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest
	rustup target add x86_64-pc-windows-gnu x86_64-unknown-linux-gnu
	rustup target add x86_64-apple-darwin aarch64-apple-darwin

# --------------------------------------------
# 2. Cross-compilation
# --------------------------------------------

cross-windows:
	@echo "$(YELLOW)➜ Building Windows x64...$(RESET)"
	# Go : Utilisation de ./cmd et suppression du cache pour éviter les erreurs de "main undeclared"
	fyne-cross windows --arch=amd64 --app-id=$(APP_ID) --app-version=$(APP_VERSION) --output=$(APP_NAME).exe --no-cache $(GO_PKG)
	# Rust
	cd $(RUST_DIR) && cargo build --release --target=x86_64-pc-windows-gnu
	# Electron (Architecture x64 forcée)
	cd $(BROWSER_DIR) && npm run build:win -- --x64

cross-linux:
	@echo "$(YELLOW)➜ Building Linux x64...$(RESET)"
	# Go : On utilise --no-cache pour forcer Linux à re-scanner le dossier ./cmd
	fyne-cross linux --arch=amd64 --app-id=$(APP_ID) --output=$(APP_NAME) --no-cache $(GO_PKG)
	# Rust
	@echo "$(YELLOW)➜ Building Rust for Linux (via Docker)...$(RESET)"
	docker run --rm --platform linux/amd64 -v $(shell pwd):/app -w /app/$(RUST_DIR) rust:1.82-bookworm /bin/bash -c "\
		apt-get update && \
		apt-get install -y libspeechd-dev pkg-config clang libclang-dev && \
		cargo build --release"
	# Electron (Architecture x64 forcée)
	cd $(BROWSER_DIR) && npm run build:linux -- --x64

cross-mac:
	@echo "$(YELLOW)➜ Building macOS (Universal)...$(RESET)"
	fyne-cross darwin --arch=amd64,arm64 --app-id=$(APP_ID) --output=$(APP_NAME) $(GO_PKG)
	cd $(RUST_DIR) && cargo build --release --target=x86_64-apple-darwin
	cd $(RUST_DIR) && cargo build --release --target=aarch64-apple-darwin
	cd $(BROWSER_DIR) && npm run build:mac

build-all: clean cross-windows cross-linux cross-mac
	@echo "$(GREEN)✔ Build multiplateforme terminé.$(RESET)"

# --------------------------------------------
# 3. Nettoyage
# --------------------------------------------
clean:
	@echo "$(YELLOW)➜ Nettoyage des binaires et fichiers temporaires...$(RESET)"
	rm -f ./main.go
	rm -rf $(BIN) fyne-cross/
	cd $(RUST_DIR) && cargo clean
	cd $(BROWSER_DIR) && rm -rf dist/
