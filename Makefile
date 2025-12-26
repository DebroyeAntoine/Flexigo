# ============================================
# Flexigo – Makefile complet multiplateforme
# Go (Fyne) + Rust TTS
# Fonctionne sous macOS / Linux / Windows
# ============================================

GO_BUILD_FLAGS=-ldflags="-s -w"
RUST_BUILD_FLAGS=--release

BIN=bin
RUST_DIR=rust/tts-rs

# --------------------------------------------
# Aide
# --------------------------------------------
.PHONY: help
help:
	@echo ""
	@echo "⌘ Flexigo Build System"
	@echo ""
	@echo "Commandes principales :"
	@echo "  make deps                → Installe toutes les dépendances"
	@echo "  make build               → Build Go + Rust natif"
	@echo "  make run                 → Lance Flexigo"
	@echo ""
	@echo "Cross-compilation Go (via fyne-cross) :"
	@echo "  make cross-linux-amd64"
	@echo "  make cross-linux-arm64"
	@echo "  make cross-windows-amd64"
	@echo "  make cross-darwin-amd64"
	@echo "  make cross-darwin-arm64"
	@echo ""
	@echo "Build complet multiplateforme :"
	@echo "  make build-all"
	@echo ""

# --------------------------------------------
# Dépendances
# --------------------------------------------
.PHONY: deps
deps:
	go mod download
	rustup update
	go install fyne.io/fyne/v2/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest
	@echo "✔ Dépendances Go/Fyne installées."

# --------------------------------------------
# Build natif (selon l'OS courant)
# --------------------------------------------
.PHONY: build-go
build-go:
	@mkdir -p $(BIN)
	go build $(GO_BUILD_FLAGS) -o $(BIN)/flexigo ./cmd/main.go
	@echo "✔ Build Go natif → $(BIN)/flexigo"

.PHONY: build-rust
build-rust:
	cd $(RUST_DIR) && cargo build $(RUST_BUILD_FLAGS)
	cp $(RUST_DIR)/target/release/tts-rs $(BIN)/flexigo-tts
	@echo "✔ Build Rust natif → $(BIN)/flexigo-tts"

.PHONY: build
build: build-go build-rust
	@echo "✔ Build natif complet terminé."

# --------------------------------------------
# Run
# --------------------------------------------
.PHONY: run
run: build
	./bin/flexigo

create-main-link:
	@if [ ! -f main.go ]; then ln -s cmd/main.go main.go; fi

remove-main-link:
	@if [ -L main.go ]; then rm main.go; fi

# --------------------------------------------
# Cross-compilation Go (via fyne-cross)
# --------------------------------------------
.PHONY: cross-linux-amd64
cross-linux-amd64: create-main-link
	fyne-cross linux --arch=amd64 --output=flexigo
	@echo "✔ Go Linux amd64 OK → fyne-cross/dist/linux-amd64"
	$(MAKE) remove-main-link

.PHONY: cross-linux-arm64
cross-linux-arm64: create-main-link
	fyne-cross linux --arch=arm64 --output=flexigo
	@echo "✔ Go Linux arm64 OK → fyne-cross/dist/linux-arm64"
	$(MAKE) remove-main-link

.PHONY: cross-windows-amd64
cross-windows-amd64: create-main-link
	fyne-cross windows --arch=amd64 --app-id=com.flexigo.app --app-version=1.0.0 --output=flexigo
	@echo "✔ Go Windows amd64 OK → fyne-cross/dist/windows-amd64"
	$(MAKE) remove-main-link

.PHONY: cross-darwin-amd64
cross-darwin-amd64: create-main-link
	fyne-cross darwin --arch=amd64 --output=flexigo
	@echo "✔ Go macOS Intel OK → fyne-cross/dist/darwin-amd64"
	$(MAKE) remove-main-link

.PHONY: cross-darwin-arm64
cross-darwin-arm64: create-main-link
	fyne-cross darwin --arch=arm64 --output=flexigo
	@echo "✔ Go macOS ARM OK → fyne-cross/dist/darwin-arm64"
	$(MAKE) remove-main-link

# ----------------------
# Rust cross-compilation
# ----------------------
.PHONY: rust-linux-amd64
rust-linux-amd64:
	rustup target add x86_64-unknown-linux-gnu
	cd $(RUST_DIR) && cargo build --release --target=x86_64-unknown-linux-gnu
	@echo "✔ Rust Linux amd64 OK"

.PHONY: rust-linux-arm64
rust-linux-arm64:
	rustup target add aarch64-unknown-linux-gnu
	cd $(RUST_DIR) && cargo build --release --target=aarch64-unknown-linux-gnu
	@echo "✔ Rust Linux ARM64 OK"

.PHONY: rust-windows-amd64
rust-windows-amd64:
	rustup target add x86_64-pc-windows-gnu
	cd $(RUST_DIR) && cargo build --release --target=x86_64-pc-windows-gnu
	@echo "✔ Rust Windows amd64 OK"

.PHONY: rust-darwin-amd64
rust-darwin-amd64:
	rustup target add x86_64-apple-darwin
	cd $(RUST_DIR) && cargo build --release --target=x86_64-apple-darwin
	@echo "✔ Rust macOS Intel OK"

.PHONY: rust-darwin-arm64
rust-darwin-arm64:
	rustup target add aarch64-apple-darwin
	cd $(RUST_DIR) && cargo build --release --target=aarch64-apple-darwin
	@echo "✔ Rust macOS ARM OK"

# --------------------------------------------
# Build all platforms (Go + Rust)
# --------------------------------------------
.PHONY: build-all
build-all: \
	cross-linux-amd64 \
	cross-linux-arm64 \
	cross-windows-amd64 \
	cross-darwin-amd64 \
	cross-darwin-arm64 \
	rust-linux-amd64 \
	rust-linux-arm64 \
	rust-windows-amd64 \
	rust-darwin-amd64 \
	rust-darwin-arm64
	@echo ""
	@echo "✔ Tous les builds multiplateformes sont terminés."
	@echo "✔ Vérifiez fyne-cross/dist/ pour les binaires Go."
	@echo "✔ Vérifiez rust/tts-rs/target/ pour les binaires Rust."

# --------------------------------------------
# Clean
# --------------------------------------------
.PHONY: clean
clean:
	rm -rf bin/
	rm -rf fyne-cross/
	cd $(RUST_DIR) && cargo clean
	@echo "✔ Nettoyage terminé."

