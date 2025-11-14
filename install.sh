#!/usr/bin/env bash
set -e

# envault installer script
# Usage: curl -sSL https://raw.githubusercontent.com/orchard9/envault/main/install.sh | bash

REPO="orchard9/envault"
BINARY_NAME="envault"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}==>${NC} $1"
}

warn() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

error() {
    echo -e "${RED}Error:${NC} $1" >&2
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os arch

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="Linux" ;;
        Darwin*)    os="Darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="Windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="x86_64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac

    echo "${os}_${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        warn "Could not fetch latest release version"
        return 1
    fi

    echo "$version"
}

# Download and install binary
install_binary() {
    local version="$1"
    local platform="$2"
    local download_url="https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}_${version}_${platform}.tar.gz"
    local tmp_dir

    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    info "Downloading ${BINARY_NAME} v${version} for ${platform}..."

    if ! curl -sSfL "$download_url" -o "$tmp_dir/${BINARY_NAME}.tar.gz"; then
        warn "Failed to download binary from ${download_url}"
        return 1
    fi

    info "Extracting archive..."
    tar -xzf "$tmp_dir/${BINARY_NAME}.tar.gz" -C "$tmp_dir"

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
    mv "$tmp_dir/${BINARY_NAME}" "$INSTALL_DIR/${BINARY_NAME}"
    chmod +x "$INSTALL_DIR/${BINARY_NAME}"

    return 0
}

# Fallback to go install
install_with_go() {
    if ! command -v go &> /dev/null; then
        error "Go is not installed and binary download failed. Please install Go or download binary manually from https://github.com/${REPO}/releases"
    fi

    info "Installing via 'go install'..."
    go install "github.com/${REPO}/cmd/${BINARY_NAME}@latest"

    local gobin="${GOBIN:-$(go env GOPATH)/bin}"
    info "Installed to ${gobin}/${BINARY_NAME}"

    # Check if GOPATH/bin is in PATH
    if [[ ":$PATH:" != *":$gobin:"* ]]; then
        warn "Add ${gobin} to your PATH to use ${BINARY_NAME}"
        echo ""
        echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "  export PATH=\"\$PATH:${gobin}\""
    fi
}

# Check if installation was successful
verify_installation() {
    local installed_path

    # Check in INSTALL_DIR first
    if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
        installed_path="$INSTALL_DIR/$BINARY_NAME"
    # Check in PATH
    elif command -v "$BINARY_NAME" &> /dev/null; then
        installed_path=$(command -v "$BINARY_NAME")
    else
        return 1
    fi

    local version
    version=$("$installed_path" version 2>/dev/null || echo "unknown")

    info "Successfully installed: $installed_path"
    echo "$version"
    return 0
}

main() {
    echo ""
    info "Installing ${BINARY_NAME}..."
    echo ""

    # Detect platform
    local platform
    platform=$(detect_platform)
    info "Detected platform: ${platform}"

    # Try to download binary from releases
    local version
    if version=$(get_latest_version); then
        info "Latest version: v${version}"

        if install_binary "$version" "$platform"; then
            echo ""
            if verify_installation; then
                echo ""
                info "Installation complete!"

                # Check if INSTALL_DIR is in PATH
                if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
                    warn "Add ${INSTALL_DIR} to your PATH to use ${BINARY_NAME}"
                    echo ""
                    echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
                    echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
                fi

                echo ""
                info "Next steps:"
                echo "  1. Install age: brew install age  (or: go install filippo.io/age/cmd/...@latest)"
                echo "  2. Run: ${BINARY_NAME} --help"
                exit 0
            fi
        fi
    fi

    # Fallback to go install
    warn "Binary download failed, trying 'go install' as fallback..."
    install_with_go

    echo ""
    if verify_installation; then
        echo ""
        info "Installation complete!"
        echo ""
        info "Next steps:"
        echo "  1. Install age: brew install age  (or: go install filippo.io/age/cmd/...@latest)"
        echo "  2. Run: ${BINARY_NAME} --help"
    else
        error "Installation verification failed"
    fi
}

main
