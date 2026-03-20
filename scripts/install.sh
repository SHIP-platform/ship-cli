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

# Download the binary
curl -sL "$DOWNLOAD_URL" -o ship

# Make it executable
chmod +x ship

echo ""
echo "Installing to /usr/local/bin (this may require your password)..."
sudo mv ship /usr/local/bin/ship

echo ""
echo "========================================"
echo "  Installation complete! 🚀             "
echo "========================================"
echo "Run 'ship tui' to get started."
