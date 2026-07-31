#!/usr/bin/env bash
set -euo pipefail

readonly RELEASE_BASE_URL="https://github.com/SHIP-platform/ship-cli/releases/latest/download"

TMP_FILE=""
DOWNLOAD_PID=""

cleanup() {
    if [[ -n "$DOWNLOAD_PID" ]] && kill -0 "$DOWNLOAD_PID" 2>/dev/null; then
        kill "$DOWNLOAD_PID" 2>/dev/null || true
        wait "$DOWNLOAD_PID" 2>/dev/null || true
    fi
    if [[ -n "$TMP_FILE" ]]; then
        rm -f "$TMP_FILE"
    fi
}

interrupt() {
    printf '\nInstallation interrupted.\n' >&2
    exit 130
}

trap cleanup EXIT
trap interrupt HUP INT TERM

detect_platform() {
    local os
    local arch

    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$os" in
        linux | darwin) ;;
        *)
            printf 'Unsupported operating system: %s\n' "$os" >&2
            return 1
            ;;
    esac

    case "$arch" in
        x86_64 | amd64) arch="amd64" ;;
        aarch64 | arm64) arch="arm64" ;;
        *)
            printf 'Unsupported architecture: %s\n' "$arch" >&2
            return 1
            ;;
    esac

    OS="$os"
    ARCH="$arch"
    BINARY_NAME="ship-${OS}-${ARCH}"
    DOWNLOAD_URL="${RELEASE_BASE_URL}/${BINARY_NAME}"
}

download_binary() {
    local curl_status
    local frame=0
    local frames='|/-\'

    if [[ -t 1 && "${TERM:-dumb}" != "dumb" ]]; then
        curl --fail --location --silent --show-error "$DOWNLOAD_URL" -o "$TMP_FILE" &
        DOWNLOAD_PID=$!

        while kill -0 "$DOWNLOAD_PID" 2>/dev/null; do
            printf '\r\033[2K[2/3] Downloading %s %s' \
                "$BINARY_NAME" "${frames:frame++%${#frames}:1}"
            sleep 0.12
        done

        if wait "$DOWNLOAD_PID"; then
            curl_status=0
        else
            curl_status=$?
        fi
        DOWNLOAD_PID=""

        if [[ "$curl_status" -ne 0 ]]; then
            printf '\r\033[2K[2/3] Download failed for %s.\n' "$BINARY_NAME" >&2
            return "$curl_status"
        fi
        printf '\r\033[2K[2/3] Downloaded %s.\n' "$BINARY_NAME"
    else
        printf '[2/3] Downloading %s...\n' "$BINARY_NAME"
        if ! curl --fail --location --silent --show-error "$DOWNLOAD_URL" -o "$TMP_FILE"; then
            printf '[2/3] Download failed for %s.\n' "$BINARY_NAME" >&2
            return 1
        fi
        printf '[2/3] Downloaded %s.\n' "$BINARY_NAME"
    fi

    if [[ ! -s "$TMP_FILE" ]]; then
        printf 'Downloaded file is empty.\n' >&2
        return 1
    fi
    chmod +x "$TMP_FILE"
}

resolve_install_dir() {
    INSTALL_DIR="/usr/local/bin"
    USE_SUDO=false

    if [[ "$EUID" -eq 0 ]]; then
        return
    fi

    if command -v sudo >/dev/null 2>&1; then
        printf '[3/3] Checking install permission (sudo may prompt)...\n'
        if sudo -v; then
            USE_SUDO=true
            return
        fi
    fi

    if [[ -d "$HOME/.local/bin" ]]; then
        INSTALL_DIR="$HOME/.local/bin"
    elif [[ -d "$HOME/bin" ]]; then
        INSTALL_DIR="$HOME/bin"
    else
        printf 'Cannot write to /usr/local/bin and no user bin directory exists.\n' >&2
        printf 'Create %s/.local/bin or run this installer with sudo access.\n' "$HOME" >&2
        return 1
    fi
}

install_binary() {
    printf '[3/3] Installing to %s/ship...\n' "$INSTALL_DIR"
    if [[ "$USE_SUDO" == true ]]; then
        sudo mv -f "$TMP_FILE" "$INSTALL_DIR/ship"
    else
        mv -f "$TMP_FILE" "$INSTALL_DIR/ship"
    fi
    TMP_FILE=""
}

main() {
    printf '%s\n' '========================================'
    printf '%s\n' '          SHIP CLI Installer            '
    printf '%s\n' '========================================'

    if ! command -v curl >/dev/null 2>&1; then
        printf 'curl is required to install SHIP CLI.\n' >&2
        return 1
    fi

    printf '[1/3] Detecting platform...\n'
    detect_platform
    printf '[1/3] Platform: %s/%s.\n' "$OS" "$ARCH"

    TMP_FILE="$(mktemp)"
    download_binary
    resolve_install_dir
    install_binary

    printf '%s\n' '========================================'
    printf '%s\n' '  Installation complete! Run: ship tui '
    printf '%s\n' '========================================'
}

main "$@"
