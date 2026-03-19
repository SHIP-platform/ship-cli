# SHIP CLI Documentation

The `ship-cli` is a powerful, interactive command-line tool built with Go. It allows developers to seamlessly interact with the SHIP Platform directly from their terminal, providing features like listing projects, viewing applications, and establishing secure port-forwarding tunnels to internal Kubernetes pods.

## Table of Contents
1. [Features](#features)
2. [Installation & Usage](#installation--usage)
3. [Architecture](#architecture)
4. [Authentication](#authentication)
5. [Port Forwarding Deep Dive](#port-forwarding-deep-dive)

---

## Features

- **Interactive TUI**: A beautiful, `k9s`-style Text User Interface built with `bubbletea`.
- **Project & App Management**: Browse your SHIP Platform projects and applications.
- **Secure Port Forwarding**: Securely tunnel traffic from your local machine to internal services (like databases) running inside the SHIP Kubernetes cluster.
- **Persistent Authentication**: Automatically saves and loads your Personal Access Token (PAT) so you don't have to enter it every time.

---

## Installation & Usage

### Building the CLI
If you have Go installed:
```bash
go build -o ship
```

If you don't have Go installed, you can build it using Docker:
```bash
docker run --rm -v $(pwd):/app -w /app golang:latest sh -c "go build -o ship"
```

### Running the CLI
To launch the interactive Text User Interface (TUI):
```bash
./ship tui
```

To run a port-forward directly without the TUI:
```bash
./ship port-forward <APP_ID> --local-port 5432 --target-port 5432
```

---

## Architecture

The CLI is structured into several modular packages:

```text
ship-cli/
├── main.go           # Entry point
├── cmd/              # Cobra commands (root, tui, port-forward)
├── api/              # REST API client for communicating with ship-api
├── ui/               # Bubbletea TUI implementation and state machine
└── config/           # Configuration and token management (~/.ship/config.json)
```

### 1. Command Layer (`cmd/`)
Built using `spf13/cobra`, this layer handles CLI argument parsing, flags, and subcommands.

### 2. API Client (`api/`)
A standard HTTP client that communicates with `https://api.ship-platform.com`. It handles fetching Projects and Applications using the user's Personal Access Token.

### 3. User Interface (`ui/`)
Built using `charmbracelet/bubbletea` (The Elm architecture for Go). It operates on a state machine:
`InputToken -> LoadProjects -> SelectProject -> LoadApps -> SelectApp -> SelectAction -> InputPorts -> PortForwarding`

### 4. Configuration (`config/`)
Handles securely storing the user's PAT in `~/.ship/config.json` so subsequent runs do not require re-authentication.

---

## Authentication

The CLI authenticates with the SHIP Platform using **Personal Access Tokens (PATs)**.

1. When you run `./ship tui` for the first time, it prompts you for your PAT.
2. The token is saved locally to `~/.ship/config.json`.
3. The CLI attaches this token as a `Bearer` token in the `Authorization` header for all REST API requests to `api.ship-platform.com`.
4. For WebSocket connections (which cannot easily send HTTP headers in all environments), the token is passed securely via the `?token=` query parameter to `console.ship-platform.com`.

---

## Port Forwarding Deep Dive

The most complex feature of the CLI is the Port Forwarding mechanism. It allows a user to securely connect to an internal, non-public pod (like a PostgreSQL database) from their local machine.

### The Flow
1. **Local Listener**: The CLI starts a local TCP server (e.g., `localhost:5432`).
2. **WebSocket Upgrade**: The CLI initiates a WebSocket connection to `wss://console.ship-platform.com/ws/portforward/<APP_ID>?port=<TARGET_PORT>&token=<PAT>`.
3. **Backend Proxy**: The `ship-api` backend verifies the token, ensures the user owns the application, and dials the internal Kubernetes Service via raw TCP.
4. **Binary Streaming**: 
   - When a local client (like `psql`) connects to the CLI, the CLI reads the raw TCP bytes.
   - It wraps these bytes in `websocket.BinaryMessage` frames and sends them to the backend.
   - The backend unwraps the frames and writes the raw bytes to the internal database pod.
   - The reverse happens for responses from the database.

### Why WebSockets?
Kubernetes Ingress controllers (like Nginx) are designed for HTTP traffic. By upgrading the connection to a WebSocket, we create a persistent, bi-directional, full-duplex tunnel through the standard HTTPS port (443) without needing to expose raw TCP ports on the cluster's edge.

### Why `console.ship-platform.com` for WebSockets?
While standard REST API calls go to `api.ship-platform.com`, the WebSocket connection specifically targets `console.ship-platform.com`. This is because the Nginx Ingress rules for the console domain are specially configured with long timeouts (`proxy-read-timeout: 3600`) to prevent database connections from dropping during idle periods.
