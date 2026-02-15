#!/bin/bash
set -e

# Download frpc binaries for embedding
# Usage: ./scripts/download-frpc.sh

VERSION="0.61.1"
BASE_URL="https://github.com/fatedier/frp/releases/download/v${VERSION}"
BINARIES_DIR="embedded/binaries"

mkdir -p "$BINARIES_DIR"

echo "Downloading frpc v${VERSION} binaries..."

# Download for different platforms
declare -A PLATFORMS=(
    ["darwin-amd64"]="frp_${VERSION}_darwin_amd64.tar.gz"
    ["darwin-arm64"]="frp_${VERSION}_darwin_arm64.tar.gz"
    ["linux-amd64"]="frp_${VERSION}_linux_amd64.tar.gz"
    ["linux-arm64"]="frp_${VERSION}_linux_arm64.tar.gz"
    ["windows-amd64"]="frp_${VERSION}_windows_amd64.zip"
)

for platform in "${!PLATFORMS[@]}"; do
    filename="${PLATFORMS[$platform]}"
    echo "  Downloading ${platform}..."
    
    # Download
    curl -sL "${BASE_URL}/${filename}" -o "/tmp/${filename}"
    
    # Extract
    if [[ $filename == *.tar.gz ]]; then
        tar -xzf "/tmp/${filename}" -C /tmp
        extracted_dir="/tmp/frp_${VERSION}_${platform}"
        cp "${extracted_dir}/frpc" "${BINARIES_DIR}/frpc-${platform}"
    elif [[ $filename == *.zip ]]; then
        unzip -q "/tmp/${filename}" -d /tmp
        extracted_dir="/tmp/frp_${VERSION}_${platform}"
        cp "${extracted_dir}/frpc.exe" "${BINARIES_DIR}/frpc-${platform}.exe"
    fi
    
    # Cleanup
    rm -f "/tmp/${filename}"
    rm -rf "$extracted_dir"
    
    echo "    ✓ ${platform}"
done

# Make binaries executable
chmod +x "${BINARIES_DIR}"/frpc-*

echo ""
echo "✓ Downloaded all frpc binaries to ${BINARIES_DIR}/"
echo ""
ls -lh "${BINARIES_DIR}/"
