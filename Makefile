# ============================================
# Flexigo – Makefile complet multiplateforme
# Go (Fyne) + Rust TTS + Electron Browser
# ============================================

GO_BUILD_FLAGS=-ldflags="-s -w"
RUST_BUILD_FLAGS=--release

BIN=bin
RUST_DIR=rust/tts-rs
BROWSER_DIR=browser
# Dossier de sortie pour electron-builder (défini dans package.json)
BROWSER_DIST=$(BIN)/browser 

# Détection de l'OS pour le build natif
ifeq ($(OS),Windows_NT)
    PLATFORM=win
    BROWSER_EXE=$(BROWSER_DIST)/win-unpacked/flexigo-browser.exe
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        PLATFORM=linux
        BROWSER_EXE=$(BROWSER_DIST)/linux-unpacked/flexigo-browser
    endif
    ifeq ($(UNAME_S),Darwin)
        PLATFORM=mac
        BROWSER_EXE=$(BROWSER_DIST)/mac/flexigo-browser.app/Contents/MacOS/flexigo-browser
    endif
endif

.PHONY: help
help:
	@echo "  make deps                → Installe Go, Rust et Node deps"
	@echo "  make build               → Build Go + Rust + Electron (natif)"
	@echo "  make build-browser       → Build Electron uniquement"
	@echo "  make build-all           → Tout build pour toutes les plateformes"

# --------------------------------------------
# Dépendances
# --------------------------------------------
.PHONY: deps
deps:
	go mod download
	rustup update
	# Installation dépendances Node/Electron
	cd $(BROWSER_DIR) && npm install
	go install fyne.io/fyne/v2/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest
	@echo "✔ Toutes les dépendances (Go/Rust/Node) sont installées."

# --------------------------------------------
# Build natif
# --------------------------------------------
.PHONY: build-go
build-go:
	@mkdir -p $(BIN)
	go build $(GO_BUILD_FLAGS) -o $(BIN)/flexigo ./cmd/main.go

.PHONY: build-rust
build-rust:
	cd $(RUST_DIR) && cargo build $(RUST_BUILD_FLAGS)
	cp $(RUST_DIR)/target/release/tts-rs $(BIN)/flexigo-tts

.PHONY: build-browser
build-browser:
	@mkdir -p $(BROWSER_DIST)
	# On lance le script build de package.json selon la plateforme détectée
	cd $(BROWSER_DIR) && npm run build:$(PLATFORM)
	@echo "✔ Build Electron terminé dans $(BROWSER_DIST)"

.PHONY: build
build: build-go build-rust build-browser
	@echo "✔ Build natif complet terminé."

# --------------------------------------------
# Cross-compilation Electron
# --------------------------------------------
# Note: electron-builder permet de cross-compiler facilement
.PHONY: cross-browser-win
cross-browser-win:
	cd $(BROWSER_DIR) && npm run build:win

.PHONY: cross-browser-linux
cross-browser-linux:
	cd $(BROWSER_DIR) && npm run build:linux

.PHONY: cross-browser-mac
cross-browser-mac:
	cd $(BROWSER_DIR) && npm run build:mac

# --------------------------------------------
# Build all platforms
# --------------------------------------------
.PHONY: build-all
build-all: \
	cross-linux-amd64 cross-linux-arm64 \
	cross-windows-amd64 \
	cross-darwin-amd64 cross-darwin-arm64 \
	rust-linux-amd64 rust-linux-arm64 \
	rust-windows-amd64 \
	rust-darwin-amd64 rust-darwin-arm64 \
	cross-browser-win cross-browser-linux cross-browser-mac
	@echo "✔ Build total terminé."

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf fyne-cross/
	cd $(RUST_DIR) && cargo clean
	rm -rf $(BROWSER_DIR)/dist
