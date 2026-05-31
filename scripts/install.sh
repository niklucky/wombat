#!/bin/sh
set -e

REPO="niklucky/wombat"
BINARY="wombat"
INSTALL_DIR="${HOME}/.local/bin"

# Get latest version from GitHub redirect (avoids API rate limits).
# Falls back to GitHub API if the redirect fails.
get_latest_version() {
    _version=$(curl -fsSI "https://github.com/${REPO}/releases/latest" 2>/dev/null \
        | grep -i '^location:' \
        | sed -E 's/.*tag\/(v[0-9]+\.[0-9]+\.[0-9]+).*/\1/')

    if [ -n "$_version" ]; then
        echo "$_version"
        return
    fi

    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null \
        | grep '"tag_name":' \
        | sed -E 's/.*"([^"]+)".*/\1/' \
        | head -n 1
}

OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Linux)     OS="linux" ;;
    Darwin)    OS="darwin" ;;
    MINGW*|CYGWIN*|MSYS*) OS="windows" ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

if [ "$OS" = "windows" ]; then
    EXT="zip"
    BINARY_NAME="${BINARY}.exe"
else
    EXT="tar.gz"
    BINARY_NAME="${BINARY}"
fi

if [ "$OS" = "linux" ] && [ "$ARCH" = "arm64" ]; then
    echo "Error: Linux arm64 binaries are not currently published."
    echo "Build from source with: go install github.com/${REPO}/cmd/${BINARY}@latest"
    exit 1
fi

echo "Detecting latest release..."
VERSION=$(get_latest_version)
if [ -z "$VERSION" ]; then
    echo "Error: unable to determine latest release version."
    exit 1
fi
echo "Latest release: ${VERSION}"

ASSET_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}-${OS}-${ARCH}.${EXT}"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${BINARY}-${OS}-${ARCH}.${EXT}..."
curl -fL# -o "${TMP_DIR}/archive.${EXT}" "$ASSET_URL"

echo "Extracting..."
if [ "$EXT" = "zip" ]; then
    unzip -q "${TMP_DIR}/archive.${EXT}" -d "$TMP_DIR"
else
    tar -xzf "${TMP_DIR}/archive.${EXT}" -C "$TMP_DIR"
fi

mkdir -p "$INSTALL_DIR"

if [ ! -w "$INSTALL_DIR" ]; then
    echo "Error: cannot write to ${INSTALL_DIR}"
    exit 1
fi

cp "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo ""
echo "${BINARY} ${VERSION} installed to ${INSTALL_DIR}/${BINARY_NAME}"

if ! command -v "$BINARY" >/dev/null 2>&1; then
    echo ""
    echo "Warning: ${INSTALL_DIR} is not on your PATH."
    echo "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
fi
