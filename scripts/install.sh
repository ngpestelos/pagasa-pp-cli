#!/usr/bin/env bash
#
# pagasa-pp-cli fleet installer — idempotent, macOS + Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/ngpestelos/pagasa-pp-cli/main/scripts/install.sh | bash
#
# Installs the pagasa-pp-cli binary via `go install`. If this machine has the
# ngpestelos fleet layout (~/src/hermes-config, ~/.claude/skills), it also
# wires the pp-pagasa skill; those steps are skipped cleanly elsewhere.
#
# Requires: git, and Go 1.21+ (GOTOOLCHAIN=auto fetches the 1.26.5 the module
# needs). On the fleet, `rebuild` provides Go; otherwise install from
# https://go.dev/dl/.

set -euo pipefail

MODULE="github.com/ngpestelos/pagasa-pp-cli"
BIN="pagasa-pp-cli"
GOBIN_DIR="${GOBIN:-$HOME/.local/bin}"

log() { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33mwarn:\033[0m %s\n' "$*" >&2; }
die() { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

OWNER_REPO="ngpestelos/pagasa-pp-cli"
mkdir -p "$GOBIN_DIR"

# --- 1. Prefer the prebuilt release binary (no local compile) ----------------
# Compiling modernc.org/sqlite is CPU-heavy — never do it on a small VPS that's
# also running services. Download the CI-built binary instead; fall back to
# `go install` only if the download can't be resolved.
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) arch="" ;;
esac

install_ok=false
if [ -n "$arch" ] && command -v curl >/dev/null 2>&1; then
  # Resolve the latest release tag (public repo, no auth needed).
  ver="$(curl -fsSL "https://api.github.com/repos/${OWNER_REPO}/releases/latest" 2>/dev/null \
    | sed -n 's/.*"tag_name": *"v\{0,1\}\([^"]*\)".*/\1/p' | head -1)"
  if [ -n "$ver" ]; then
    tarball="pagasa-pp-cli_${ver}_${os}_${arch}.tar.gz"
    url="https://github.com/${OWNER_REPO}/releases/download/v${ver}/${tarball}"
    log "Downloading prebuilt $BIN v$ver ($os/$arch)"
    tmp="$(mktemp -d)"
    if curl -fsSL "$url" -o "$tmp/$tarball" 2>/dev/null && tar -xzf "$tmp/$tarball" -C "$GOBIN_DIR" 2>/dev/null; then
      chmod +x "$GOBIN_DIR/pagasa-pp-cli" "$GOBIN_DIR/pagasa-pp-mcp" 2>/dev/null || true
      install_ok=true
    else
      warn "prebuilt download failed ($url); will try building from source."
    fi
    rm -rf "$tmp"
  fi
fi

# --- 2. Fallback: build from source (go install) -----------------------------
if [ "$install_ok" != true ]; then
  command -v go >/dev/null 2>&1 || die "No prebuilt binary available and Go not on PATH. Install Go 1.21+ (https://go.dev/dl/) or check the release assets."
  warn "Building from source — this compiles modernc.org/sqlite and is CPU-heavy."
  # A brand-new public module can 500 on sum.golang.org while the checksum DB
  # indexes it; retry a few times before giving up.
  for attempt in 1 2 3; do
    if GOTOOLCHAIN=auto GOBIN="$GOBIN_DIR" go install "${MODULE}/cmd/${BIN}@latest" 2>/tmp/pagasa-install.err; then
      install_ok=true
      break
    fi
    if grep -q "sum.golang.org" /tmp/pagasa-install.err 2>/dev/null; then
      warn "checksum DB not ready (attempt $attempt/3); retrying in 10s"
      sleep 10
    else
      cat /tmp/pagasa-install.err >&2
      break
    fi
  done
  rm -f /tmp/pagasa-install.err
fi
[ "$install_ok" = true ] || die "install failed (neither prebuilt download nor go install worked)."

# --- 3. Verify ---------------------------------------------------------------
BIN_PATH="$GOBIN_DIR/$BIN"
[ -x "$BIN_PATH" ] || die "binary not found at $BIN_PATH after install."
log "Installed: $("$BIN_PATH" --version 2>/dev/null || echo "$BIN (version unknown)")"
"$BIN_PATH" now --dry-run >/dev/null 2>&1 && log "Smoke check passed (now --dry-run)." || warn "dry-run smoke check failed; the binary installed but check PATH/network."

case ":$PATH:" in
  *":$GOBIN_DIR:"*) : ;;
  *) warn "$GOBIN_DIR is not on your PATH — add it so agents can find $BIN." ;;
esac

# --- 4. Fleet skill wiring (guarded; no-op off the fleet) --------------------
HERMES_CONFIG="$HOME/src/hermes-config"
if [ -d "$HERMES_CONFIG/skills/pp-pagasa" ]; then
  log "Refreshing hermes-config (pp-pagasa skill)"
  git -C "$HERMES_CONFIG" pull --rebase --autostash -q 2>/dev/null || warn "could not pull hermes-config; skill may be stale."
  # Claude Code discovers skills via ~/.claude/skills symlinks; Hermes reads
  # hermes-config directly and needs no symlink.
  if [ -d "$HOME/.claude/skills" ]; then
    ln -sfn "$HERMES_CONFIG/skills/pp-pagasa" "$HOME/.claude/skills/pp-pagasa"
    log "Linked pp-pagasa into ~/.claude/skills (Claude Code)."
  fi
  log "pp-pagasa skill ready. Restart the Hermes gateway/session to load it."
else
  log "No fleet skill layout here — binary only. (pp-pagasa lives in ~/src/hermes-config on the fleet.)"
fi

log "Done."
