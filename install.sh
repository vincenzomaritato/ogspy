#!/usr/bin/env bash
# Usage: curl -fsSL https://raw.githubusercontent.com/vincenzomaritato/ogspy/main/install.sh | bash
set -euo pipefail

REPO="vincenzomaritato/ogspy"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)            ARCH=amd64 ;;
  aarch64 | arm64)   ARCH=arm64 ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

VERSION=${1:-$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
            | grep -Po '"tag_name": "\Kv[^"]+')}

TAR="ogspy_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TAR}"

echo "Downloading ${URL}"
curl -L --proto '=https' --tlsv1.2 -o "${TAR}" "${URL}"
curl -L --proto '=https' --tlsv1.2 -o "${TAR}.sig" "${URL}.sig"

echo "Verifying signature..."
cosign verify-blob -in "${TAR}" -sig "${TAR}.sig"

echo "Installing to /usr/local/bin (sudo may prompt for password)..."
sudo tar -C /usr/local/bin -xzvf "${TAR}" ogspy
echo "ogspy successfully installed."