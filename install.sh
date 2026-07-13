#!/usr/bin/env sh
# Installer for chirp: https://github.com/stinmark/chirp
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/stinmark/chirp/main/install.sh | sh
#
# Optional env vars:
#   CHIRP_INSTALL_DIR   Where to install the binary (default: /usr/local/bin,
#                       or $HOME/.local/bin if that's not writable)
#   CHIRP_VERSION       Install a specific tag instead of the latest release,
#                       e.g. CHIRP_VERSION=v0.3.0

set -eu

REPO="stinmark/chirp"
BIN_NAME="chirp"

info() { printf '\033[1;34m==>\033[0m %s\n' "$1"; }
warn() { printf '\033[1;33mwarning:\033[0m %s\n' "$1"; }
error() {
  printf '\033[1;31merror:\033[0m %s\n' "$1" >&2
  exit 1
}

detect_os() {
  case "$(uname -s)" in
  Linux) echo "linux" ;;
  Darwin) error "chirp does not currently publish macOS builds. See https://github.com/${REPO} for details." ;;
  MINGW* | MSYS* | CYGWIN*) error "Detected Windows. Please download the .zip from https://github.com/${REPO}/releases instead of using install.sh." ;;
  *) error "Unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
  x86_64 | amd64) echo "amd64" ;;
  aarch64 | arm64) echo "arm64" ;;
  *) error "Unsupported architecture: $(uname -m)" ;;
  esac
}

pick_install_dir() {
  if [ -n "${CHIRP_INSTALL_DIR:-}" ]; then
    echo "$CHIRP_INSTALL_DIR"
    return
  fi
  if [ -w "/usr/local/bin" ]; then
    echo "/usr/local/bin"
  else
    echo "$HOME/.local/bin"
  fi
}

latest_version() {
  # Queries the UI redirect instead of the API to completely avoid 403 rate limits
  curl -f -S -L -I -o /dev/null -w "%{url_effective}" "https://github.com/${REPO}/releases/latest" |
    awk -F'/' '{print $NF}'
}

main() {
  command -v curl >/dev/null 2>&1 || error "curl is required but not installed."
  command -v tar >/dev/null 2>&1 || error "tar is required but not installed."

  os="$(detect_os)"
  arch="$(detect_arch)"

  version="${CHIRP_VERSION:-}"
  if [ -z "$version" ]; then
    info "Looking up the latest chirp release..."
    version="$(latest_version)"
    [ -n "$version" ] || error "Could not determine the latest release. Set CHIRP_VERSION to install a specific tag."
  fi

  archive="chirp_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${version}/${archive}"

  install_dir="$(pick_install_dir)"
  mkdir -p "$install_dir"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT

  info "Downloading chirp ${version} for ${os}/${arch}..."
  curl -fsSL "$url" -o "${tmp_dir}/${archive}" ||
    error "Failed to download ${url}. Check that a release exists for your platform."

  info "Installing to ${install_dir}..."
  tar -xzf "${tmp_dir}/${archive}" -C "$tmp_dir" "$BIN_NAME"
  chmod +x "${tmp_dir}/${BIN_NAME}"
  mv "${tmp_dir}/${BIN_NAME}" "${install_dir}/${BIN_NAME}"

  info "chirp ${version} installed to ${install_dir}/${BIN_NAME}"

  case ":$PATH:" in
  *":${install_dir}:"*) ;;
  *) warn "${install_dir} is not on your PATH. Add this to your shell profile:
    export PATH=\"${install_dir}:\$PATH\"" ;;
  esac

  "${install_dir}/${BIN_NAME}" --help >/dev/null 2>&1 || true
  info "Run 'chirp' to open the dashboard, or 'chirp --run-daemon' to start the background scheduler."
}

main "$@"
