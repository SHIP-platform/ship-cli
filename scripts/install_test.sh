#!/usr/bin/env bash
set -euo pipefail

readonly ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly INSTALLER="$ROOT_DIR/scripts/install.sh"
TEST_ROOT="$(mktemp -d)"

cleanup() {
    rm -rf "$TEST_ROOT"
}
trap cleanup EXIT

fail() {
    printf 'FAIL: %s\n' "$1" >&2
    exit 1
}

assert_contains() {
    local output="$1"
    local expected="$2"
    [[ "$output" == *"$expected"* ]] || fail "expected output to contain: $expected"
}

create_mocks() {
    local mock_dir="$1"
    mkdir -p "$mock_dir"

    cat >"$mock_dir/uname" <<'EOF'
#!/usr/bin/env bash
case "${1:-}" in
    -s) printf 'Linux\n' ;;
    -m) printf 'x86_64\n' ;;
    *) exit 1 ;;
esac
EOF

    cat >"$mock_dir/curl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
output=""
while [[ "$#" -gt 0 ]]; do
    case "$1" in
        -o | --output)
            output="$2"
            shift 2
            ;;
        *) shift ;;
    esac
done
sleep "${MOCK_CURL_DELAY:-0}"
if [[ "${MOCK_CURL_FAIL:-0}" == "1" ]]; then
    printf 'curl: mocked download failure\n' >&2
    exit 22
fi
printf '#!/usr/bin/env sh\nprintf "SHIP CLI mock\\n"\n' >"$output"
EOF

    cat >"$mock_dir/sudo" <<'EOF'
#!/usr/bin/env bash
exit 1
EOF

    cat >"$mock_dir/mktemp" <<'EOF'
#!/usr/bin/env bash
: >"$MOCK_TMP_FILE"
printf '%s\n' "$MOCK_TMP_FILE"
EOF

    chmod +x "$mock_dir/uname" "$mock_dir/curl" "$mock_dir/sudo" "$mock_dir/mktemp"
}

run_non_tty_success() {
    local case_dir="$TEST_ROOT/non-tty"
    local mock_dir="$case_dir/mocks"
    local home_dir="$case_dir/home"
    local output
    mkdir -p "$home_dir/.local/bin"
    create_mocks "$mock_dir"

    output="$(env \
        PATH="$mock_dir:$PATH" \
        HOME="$home_dir" \
        MOCK_TMP_FILE="$case_dir/download.tmp" \
        bash "$INSTALLER" 2>&1)"

    assert_contains "$output" "[1/3] Platform: linux/amd64."
    assert_contains "$output" "[2/3] Downloading ship-linux-amd64..."
    assert_contains "$output" "[2/3] Downloaded ship-linux-amd64."
    assert_contains "$output" "[3/3] Checking install permission (sudo may prompt)..."
    assert_contains "$output" "[3/3] Installing to $home_dir/.local/bin/ship..."
    assert_contains "$output" "Installation complete! Run: ship tui"
    [[ "$output" != *$'\r'* ]] || fail "non-TTY output contains carriage returns"
    [[ -x "$home_dir/.local/bin/ship" ]] || fail "installer did not create an executable"
    [[ ! -e "$case_dir/download.tmp" ]] || fail "temporary download was not moved or cleaned"
}

run_tty_success() {
    local case_dir="$TEST_ROOT/tty"
    local mock_dir="$case_dir/mocks"
    local home_dir="$case_dir/home"
    local spinner_count
    local transcript="$case_dir/transcript.log"
    mkdir -p "$home_dir/.local/bin"
    create_mocks "$mock_dir"

    if script --version 2>&1 | grep -qi 'util-linux'; then
        script -q -e -c \
            "env PATH='$mock_dir:$PATH' HOME='$home_dir' TERM=xterm MOCK_CURL_DELAY=0.35 MOCK_TMP_FILE='$case_dir/download.tmp' bash '$INSTALLER'" \
            "$transcript" >/dev/null
    else
        script -q "$transcript" env \
            PATH="$mock_dir:$PATH" \
            HOME="$home_dir" \
            TERM=xterm \
            MOCK_CURL_DELAY=0.35 \
            MOCK_TMP_FILE="$case_dir/download.tmp" \
            bash "$INSTALLER" >/dev/null
    fi

    grep -Fq "[2/3] Downloading ship-linux-amd64" "$transcript" || \
        fail "TTY output did not show the spinner stage"
    spinner_count="$(tr '\r' '\n' <"$transcript" | grep -Fc '[2/3] Downloading ship-linux-amd64')"
    [[ "$spinner_count" -ge 2 ]] || fail "TTY output did not render multiple spinner frames"
    grep -Fq "[2/3] Downloaded ship-linux-amd64." "$transcript" || \
        fail "TTY output did not show download completion"
    grep -q $'\r' "$transcript" || fail "TTY output did not animate with carriage returns"
    [[ -x "$home_dir/.local/bin/ship" ]] || fail "TTY installer did not create an executable"
}

run_download_failure() {
    local case_dir="$TEST_ROOT/failure"
    local mock_dir="$case_dir/mocks"
    local home_dir="$case_dir/home"
    local output
    mkdir -p "$home_dir/.local/bin"
    create_mocks "$mock_dir"

    if output="$(env \
        PATH="$mock_dir:$PATH" \
        HOME="$home_dir" \
        MOCK_CURL_FAIL=1 \
        MOCK_TMP_FILE="$case_dir/download.tmp" \
        bash "$INSTALLER" 2>&1)"; then
        fail "installer succeeded after a mocked download failure"
    fi

    assert_contains "$output" "Download failed for ship-linux-amd64."
    [[ ! -e "$home_dir/.local/bin/ship" ]] || fail "failed download installed a binary"
    [[ ! -e "$case_dir/download.tmp" ]] || fail "failed download left a temporary file"
}

run_non_tty_success
run_tty_success
run_download_failure
printf 'PASS: installer progress and failure behavior\n'
