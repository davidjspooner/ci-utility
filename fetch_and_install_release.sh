#!/bin/bash

set -uo pipefail

# Check for required tools
for tool in curl jq zip; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "Error: $tool is required but not installed." >&2
        exit 1
    fi
done

if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    CURL_AUTH=(-H "Authorization: token $GITHUB_TOKEN")
else
    CURL_AUTH=()
fi

VERSION="${1:-}"

# if version =="" or version == "latest" then fetch the latest version
if [[ "$VERSION" == "latest" ]]; then
    VERSION=""
fi
if [[ -z "$VERSION" ]]; then
    echo "No version specified. Fetching latest..."
    VERSION=$(curl -fsSL "${CURL_AUTH[@]}" "https://api.github.com/repos/davidjspooner/ci-utility/releases/latest" | jq -r .tag_name)
    if [[ -z "$VERSION" ]]; then
        echo "Failed to fetch latest version from GitHub API"
        exit 1
    fi
    echo "Latest version is [$VERSION]"
fi

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "QUERY_URL: ${CURL_AUTH[@]} https://api.github.com/repos/davidjspooner/ci-utility/releases/tags/${VERSION}"
# List all the assets for the release
ASSETS=$(curl -fsSL "${CURL_AUTH[@]}" "https://api.github.com/repos/davidjspooner/ci-utility/releases/tags/${VERSION}" )
if [[ -z "$ASSETS" ]]; then
    echo "No assets found for version $VERSION"
    exit 1
fi
ASSET_NAME=ci-utility-${OS}-${ARCH}.zip
# find the url for the asset
ASSET_URL=$(echo "$ASSETS" | jq -r ".assets[] | select(.name == \"$ASSET_NAME\") | .url")
if [[ -z "$ASSET_URL" ]]; then
    echo "No asset found for version $VERSION with name $ASSET_NAME"
    exit 1
fi

echo "ASSET_URL: $ASSET_URL"

INSTALL_PATH="/usr/local/bin/ci-utility"

echo "Downloading from $ASSET_URL"
TMP_ZIP_FILE="$(mktemp)"
trap 'rm -f "$TMP_ZIP_FILE"' EXIT

if ! curl -fsSL "${CURL_AUTH[@]}" -H "Accept: application/octet-stream" "$ASSET_URL" -o "$TMP_ZIP_FILE"; then
    echo "Download failed from $ASSET_URL"
    exit 1
fi

UNZIP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_ZIP_FILE" "$UNZIP_DIR"' EXIT

if ! unzip -o "$TMP_ZIP_FILE" -d "$UNZIP_DIR"; then
    echo "Failed to unzip $TMP_ZIP_FILE"
    exit 1
fi

# Find the binary (assume it's named ci-utility)
FOUND_BIN="$UNZIP_DIR/ci-utility"
if [[ ! -f "$FOUND_BIN" ]]; then
    echo "ci-utility binary not found in zip archive"
    exit 1
fi

chmod +x "$FOUND_BIN"

if [ -w "$(dirname "$INSTALL_PATH")" ]; then
    mv "$FOUND_BIN" "$INSTALL_PATH"
else
    sudo mv "$FOUND_BIN" "$INSTALL_PATH"
fi

echo -e "ci-utility installed to $INSTALL_PATH\n"
"$INSTALL_PATH" version 