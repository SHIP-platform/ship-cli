#!/bin/bash

# Build script for SHIP CLI cross-compilation

set -e

# Directory where the binaries will be saved
OUTPUT_DIR="build"

# Clean previous builds
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

echo "Building SHIP CLI..."

# Linux
echo "Building for Linux (amd64)..."
docker run --rm -v $(pwd):/app -w /app golang:latest sh -c "GOOS=linux GOARCH=amd64 go build -o $OUTPUT_DIR/ship-linux-amd64"

echo "Building for Linux (arm64)..."
docker run --rm -v $(pwd):/app -w /app golang:latest sh -c "GOOS=linux GOARCH=arm64 go build -o $OUTPUT_DIR/ship-linux-arm64"

# macOS
echo "Building for macOS (amd64/Intel)..."
docker run --rm -v $(pwd):/app -w /app golang:latest sh -c "GOOS=darwin GOARCH=amd64 go build -o $OUTPUT_DIR/ship-darwin-amd64"

echo "Building for macOS (arm64/Apple Silicon)..."
docker run --rm -v $(pwd):/app -w /app golang:latest sh -c "GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_DIR/ship-darwin-arm64"

# Windows
echo "Building for Windows (amd64)..."
docker run --rm -v $(pwd):/app -w /app golang:latest sh -c "GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/ship-windows-amd64.exe"

# Fix permissions since docker runs as root
docker run --rm -v $(pwd):/app -w /app alpine chown -R $(id -u):$(id -g) $OUTPUT_DIR

echo ""
echo "Build complete! Binaries are located in the '$OUTPUT_DIR' directory:"
ls -lh "$OUTPUT_DIR"
