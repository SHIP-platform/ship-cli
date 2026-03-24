#!/bin/bash
set -e

# SHIP CLI Installation Script

echo "========================================"
echo "      Downloading SHIP CLI...           "
echo "========================================"

# Detect OS
OS="$(uname -s | tr A-Z a-z)"
# Detect Architecture
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Define the binary name based on OS and ARCH
BINARY_NAME="ship-${OS}-${ARCH}"

# Fetch the latest release from GitHub
DOWNLOAD_URL="https://github.com/SHIP-platform/ship-cli/releases/latest/download/${BINARY_NAME}"

echo "Detected OS: $OS"
echo "Detected Architecture: $ARCH"
echo "Fetching from: $DOWNLOAD_URL"
echo ""

# Download the binary to a temporary file first
TMP_FILE=$(mktemp)
curl -sL "$DOWNLOAD_URL" -o "$TMP_FILE"

# Make it executable
chmod +x "$TMP_FILE"

echo ""
# Check if /usr/local/bin exists and is in PATH, fallback to ~/.local/bin or ~/bin
INSTALL_DIR="/usr/local/bin"

# If the user doesn't have sudo rights and isn't root, we might need to install to a local bin
if [ "$EUID" -ne 0 ] && ! sudo -v >/dev/null 2>&1; then
    if [ -d "$HOME/.local/bin" ]; then
        INSTALL_DIR="$HOME/.local/bin"
    elif [ -d "$HOME/bin" ]; then
        INSTALL_DIR="$HOME/bin"
    else
        echo "Requires sudo privileges to install to /usr/local/bin. Please run with sudo."
        exit 1
    fi
    echo "Installing locally to $INSTALL_DIR..."
    mv "$TMP_FILE" "$INSTALL_DIR/ship"
else
    echo "Installing to $INSTALL_DIR (this may require your password)..."
    # Use -f to forcefully overwrite existing binary
    sudo mv -f "$TMP_FILE" "$INSTALL_DIR/ship"
fi

echo ""
echo "========================================"
echo "  Installation complete! 🚀             "
echo "========================================"
echo "Run 'ship --help' to get started."
