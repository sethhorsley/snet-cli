#!/bin/bash
# Update Homebrew formula with version and SHA256 checksums from GitHub release
#
# Usage: ./scripts/update-homebrew-formula.sh v1.0.0

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION="$1"
VERSION_NO_V="${VERSION#v}"  # Remove 'v' prefix
TAP_DIR="${2:-/Users/send16/files/sethhorsley/homebrew-tap}"
FORMULA_FILE="$TAP_DIR/Formula/snet.rb"
REPO="sethhorsley/snet-cli"

echo "Updating Homebrew formula for version $VERSION..."

# Download SHA256 files from GitHub release
BASE_URL="https://github.com/$REPO/releases/download/$VERSION"

echo "Fetching SHA256 checksums from GitHub release..."

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Download SHA256 files
curl -sSL "$BASE_URL/snet-darwin-arm64.tar.gz.sha256" -o "$TEMP_DIR/darwin-arm64.sha256"
curl -sSL "$BASE_URL/snet-darwin-amd64.tar.gz.sha256" -o "$TEMP_DIR/darwin-amd64.sha256"
curl -sSL "$BASE_URL/snet-linux-arm64.tar.gz.sha256" -o "$TEMP_DIR/linux-arm64.sha256"
curl -sSL "$BASE_URL/snet-linux-amd64.tar.gz.sha256" -o "$TEMP_DIR/linux-amd64.sha256"

# Extract SHA256 hashes (first field from sha256sum output)
DARWIN_ARM64_SHA=$(awk '{print $1}' "$TEMP_DIR/darwin-arm64.sha256")
DARWIN_AMD64_SHA=$(awk '{print $1}' "$TEMP_DIR/darwin-amd64.sha256")
LINUX_ARM64_SHA=$(awk '{print $1}' "$TEMP_DIR/linux-arm64.sha256")
LINUX_AMD64_SHA=$(awk '{print $1}' "$TEMP_DIR/linux-amd64.sha256")

echo "Darwin ARM64 SHA256: $DARWIN_ARM64_SHA"
echo "Darwin AMD64 SHA256: $DARWIN_AMD64_SHA"
echo "Linux ARM64 SHA256:  $LINUX_ARM64_SHA"
echo "Linux AMD64 SHA256:  $LINUX_AMD64_SHA"

# Update formula file
echo "Updating $FORMULA_FILE..."

# Write the formula
cat > "$FORMULA_FILE" <<EOF
class Snet < Formula
  desc "Secure HTTPS tunnels from localhost to public URLs"
  homepage "https://github.com/sethhorsley/snet-cli"
  version "$VERSION_NO_V"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/sethhorsley/snet-cli/releases/download/$VERSION/snet-darwin-arm64.tar.gz"
      sha256 "$DARWIN_ARM64_SHA"
    else
      url "https://github.com/sethhorsley/snet-cli/releases/download/$VERSION/snet-darwin-amd64.tar.gz"
      sha256 "$DARWIN_AMD64_SHA"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/sethhorsley/snet-cli/releases/download/$VERSION/snet-linux-arm64.tar.gz"
      sha256 "$LINUX_ARM64_SHA"
    else
      url "https://github.com/sethhorsley/snet-cli/releases/download/$VERSION/snet-linux-amd64.tar.gz"
      sha256 "$LINUX_AMD64_SHA"
    end
  end

  def install
    bin.install "snet"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/snet version")
  end
end
EOF

echo "✅ Homebrew formula updated successfully!"
echo ""
echo "Next steps:"
echo "  cd $TAP_DIR"
echo "  git diff Formula/snet.rb"
echo "  git add Formula/snet.rb"
echo "  git commit -m 'Update snet to $VERSION'"
echo "  git push origin main"
