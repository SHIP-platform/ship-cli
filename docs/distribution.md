# SHIP CLI Distribution Guide

This guide explains how to package and distribute the `ship-cli` tool so users can easily download and install it.

## 1. Cross-Compilation

Go makes it incredibly easy to compile binaries for multiple operating systems and architectures. You can build the CLI for Windows, macOS, and Linux from a single machine.

Run these commands in your `ship-cli` directory:

```bash
# Build for Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o build/ship-linux-amd64

# Build for Linux (arm64)
GOOS=linux GOARCH=arm64 go build -o build/ship-linux-arm64

# Build for macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o build/ship-darwin-amd64

# Build for macOS (Apple Silicon / M1 / M2)
GOOS=darwin GOARCH=arm64 go build -o build/ship-darwin-arm64

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o build/ship-windows-amd64.exe
```

## 2. Hosting the Binaries

You need a place to host these binaries so users can download them. Since you already have the SHIP Platform infrastructure, you have a few good options:

### Option A: GitHub Releases (Recommended)
If your code is hosted on GitHub, you can create a "Release" and upload the compiled binaries as assets. This is the standard open-source way.

### Option B: Host on `api.ship-platform.com`
You can create a new route in your `ship-api` (e.g., `GET /api/cli/download/{os}/{arch}`) that serves the binary files directly from a cloud storage bucket (like AWS S3 or MinIO) or from the container's local filesystem.

## 3. Installation Scripts

The supported installer is [`scripts/install.sh`](../scripts/install.sh). The
Console serves that file at `https://console.ship-platform.com/install.sh`, and
the script resolves the latest GitHub Release asset for the detected Linux or
macOS architecture.

The installer must keep these user-facing guarantees:

- display platform detection, download, and installation stages;
- animate the download only when stdout is an interactive terminal;
- keep redirected/CI output line-based and free of terminal control noise;
- fail on HTTP errors and remove partial temporary files;
- fall back to an existing `~/.local/bin` or `~/bin` when sudo is unavailable.

Verify installer changes without network or system-directory writes:

```bash
bash -n scripts/install.sh scripts/install_test.sh
bash scripts/install_test.sh
```

Users can then install your CLI with a single command:
```bash
curl -sL https://console.ship-platform.com/install.sh | bash
```

## 4. Console Integration (Blog/Docs)

To promote the CLI in your `ship-console`, you should create a dedicated "CLI & Tools" page or a blog post.

Here is a markdown template you can use for your console docs:

```markdown
# Introducing the SHIP CLI

We are excited to announce the `ship` CLI! Manage your projects, view applications, and securely tunnel into your internal databases directly from your terminal.

## Installation

Install the CLI with a single command:

\`\`\`bash
curl -sL https://console.ship-platform.com/install.sh | bash
\`\`\`

*(Windows users can download the `.exe` directly from our [Releases page](#)).*

## Getting Started

Launch the interactive terminal UI:

\`\`\`bash
ship tui
\`\`\`

1. **Authenticate**: Paste your Personal Access Token (PAT) when prompted.
2. **Navigate**: Use your arrow keys to select a Project, then an Application.
3. **Port Forward**: Select "Start Port Forward" to securely connect to your internal databases (like PostgreSQL or MongoDB) from your local machine.

## Features

- **Multi-Port Forwarding**: You can now port-forward multiple applications simultaneously! The CLI runs them in the background.
- **Visual Indicators**: Applications with active port-forwards are marked with a ⚡ icon in the list.
- **Persistent Login**: Your token is securely saved, so you only need to log in once.
```
