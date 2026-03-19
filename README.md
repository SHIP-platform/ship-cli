<div align="center">
  <h1>🚢 SHIP CLI</h1>
  <p><strong>The official command-line interface for the SHIP Platform.</strong></p>
</div>

The `ship` CLI is a powerful, interactive terminal tool designed to help developers manage their projects, view applications, and securely tunnel traffic into internal Kubernetes pods without ever leaving the terminal.

---

## ✨ Features

- **Interactive TUI**: A beautiful, `k9s`-style Text User Interface built with `bubbletea`.
- **Project & App Management**: Browse your SHIP Platform projects and applications in real-time.
- **Secure Port Forwarding**: Securely tunnel TCP traffic from your local machine to internal services (like PostgreSQL, MongoDB, Redis) running inside the SHIP Kubernetes cluster.
- **Multi-Port Forwarding**: Run multiple port-forwards simultaneously in the background.
- **Persistent Authentication**: Automatically saves and loads your Personal Access Token (PAT) so you only have to log in once.

---

## 🚀 Quick Start

### 1. Installation

You can install the CLI with a single command. This script will automatically detect your operating system (Linux or macOS) and architecture, download the correct binary, and install it to `/usr/local/bin/ship`.

```bash
curl -sL https://console.ship-platform.com/install.sh | bash
```

*(If you are on Windows, you can download the `.exe` directly from the [Releases](https://github.com/SHIP-platform/ship-cli/releases) page).*

### 2. Launch the TUI

Start the interactive terminal UI by running:

```bash
ship tui
```

### 3. Authenticate

The first time you run the CLI, you will be prompted to enter your **Personal Access Token (PAT)**. 
You can generate a new PAT by logging into the SHIP Console and navigating to your Account Settings.

Once entered, your token is securely saved to `~/.ship/config.json`.

---

## 📖 Usage Guide

### Navigating the TUI

The TUI is designed to be fully navigable using your keyboard:

- `↑` / `↓`: Move up and down through lists.
- `Enter`: Select a Project, Application, or Action.
- `esc` / `q`: Go back to the previous screen.
- `L` (Shift+L): Log out and clear your saved token.
- `ctrl+c`: Force quit the application and safely close all active port-forwards.

### How to Port Forward

1. Select a **Project** from the list.
2. Select an **Application** (e.g., your PostgreSQL database).
3. Select **Start Port Forward**.
4. Enter a **Local Port** (e.g., `5432`). This is the port you will connect to on your own computer.
5. Enter a **Target Port** (e.g., `5432`). This is the port the application is listening on inside the cluster.

Once started, the CLI will run the tunnel in the background. You will see a ⚡ icon next to the application name indicating that it is actively being forwarded. 

You can now connect to your database using your favorite local tool:
```bash
psql -h localhost -p 5432 -U postgres
```

### Direct Commands (Non-Interactive)

If you prefer to use standard CLI commands (useful for scripts or CI/CD), you can bypass the TUI entirely:

```bash
# Forward local port 5432 to the pod's port 5432
ship port-forward <APP_ID> --local-port 5432 --target-port 5432
```

---

## 🛠️ Development

If you want to contribute to the CLI or build it from source, you will need Go 1.22+ installed.

### Build from source
```bash
git clone https://github.com/SHIP-platform/ship-cli.git
cd ship-cli
go build -o ship
```

### Cross-Compilation
We provide a script to compile the CLI for all major platforms (Linux, macOS, Windows) using Docker.

```bash
./scripts/build-all.sh
```
The compiled binaries will be placed in the `build/` directory.

---

## 📄 License

This project is licensed under the MIT License.
