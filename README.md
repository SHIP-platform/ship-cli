<div align="center">
  <h1>🚢 SHIP CLI</h1>
  <p><strong>The official command-line interface for the SHIP Platform.</strong></p>
</div>

The `ship` CLI is an interactive terminal tool to manage projects, view applications, and securely tunnel traffic into internal Kubernetes pods.

## 🚀 Quick Start

### 1. Installation

Install the CLI to `/usr/local/bin/ship` with a single command (Linux/macOS):

```bash
curl -sL https://console.ship-platform.com/install.sh | bash
```
*(Windows users: download the `.exe` from [Releases](https://github.com/SHIP-platform/ship-cli/releases)).*

### 2. Launch & Authenticate

Start the interactive UI:

```bash
ship tui
```
*The first time you run it, paste your **Personal Access Token (PAT)** from the SHIP Console. It will be securely saved for future use.*

---

## 📖 How to Port Forward

Use the CLI to securely connect to internal databases (PostgreSQL, MongoDB, etc.) from your local machine:

1. Use `↑` / `↓` and `Enter` to select a **Project** ➔ **Application**.
2. Select **Start Port Forward**.
3. Enter a **Local Port** (e.g., `5432`) and a **Target Port** (e.g., `5432`).

A ⚡ icon will appear next to the app, indicating the tunnel is running in the background. You can now connect using your local tools:
```bash
psql -h localhost -p 5432 -U postgres
```

### Shortcuts
- `esc` / `q`: Go back
- `L` (Shift+L): Log out
- `ctrl+c`: Quit and close all tunnels

### Direct Command
Bypass the UI for scripts:
```bash
ship port-forward <APP_ID> --local-port 5432 --target-port 5432
```

---

## 🛠️ Development

Requires Go 1.22+.

```bash
git clone https://github.com/SHIP-platform/ship-cli.git
cd ship-cli
go build -o ship

# Cross-compile for all platforms (requires Docker)
./scripts/build-all.sh
```
