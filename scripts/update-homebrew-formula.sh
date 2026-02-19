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
FORMULA_FILE="homebrew/snet.rb"
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

# Backup original file
cp "$FORMULA_FILE" "$FORMULA_FILE.bak"

# Update version and SHA256 values
sed -i.tmp "s/version \".*\"/version \"$VERSION_NO_V\"/" "$FORMULA_FILE"
sed -i.tmp "s|releases/download/v[^/]*/|releases/download/$VERSION/|g" "$FORMULA_FILE"

# Update SHA256 values for each platform
sed -i.tmp "/if Hardware::CPU.arm?/,/sha256/ { s/sha256 \".*\"/sha256 \"$DARWIN_ARM64_SHA\"/; }" "$FORMULA_FILE"
sed -i.tmp "/else.*# Darwin AMD64/,/sha256/ { s/sha256 \".*\"/sha256 \"$DARWIN_AMD64_SHA\"/; }" "$FORMULA_FILE"

# Use a more precise approach with line-by-line processing
cat "$FORMULA_FILE.bak" | awk -v v="$VERSION_NO_V" \
    -v darwin_arm64="$DARWIN_ARM64_SHA" \
    -v darwin_amd64="$DARWIN_AMD64_SHA" \
    -v linux_arm64="$LINUX_ARM64_SHA" \
    -v linux_amd64="$LINUX_AMD64_SHA" \
    -v version="$VERSION" '
BEGIN { platform=""; }
/^  version / { print "  version \"" v "\""; next; }
/on_macos/ { platform="macos"; }
/on_linux/ { platform="linux"; }
/if Hardware::CPU.arm\?/ { is_arm=1; }
/else/ { is_arm=0; }
/releases\/download/ {
    if (platform == "macos" && is_arm) {
        print "      url \"https://github.com/sethhorsley/snet-cli/releases/download/" version "/snet-darwin-arm64.tar.gz\"";
    } else if (platform == "macos" && !is_arm) {
        print "      url \"https://github.com/sethhorsley/snet-cli/releases/download/" version "/snet-darwin-amd64.tar.gz\"";
    } else if (platform == "linux" && is_arm) {
        print "      url \"https://github.com/sethhorsley/snet-cli/releases/download/" version "/snet-linux-arm64.tar.gz\"";
    } else if (platform == "linux" && !is_arm) {
        print "      url \"https://github.com/sethhorsley/snet-cli/releases/download/" version "/snet-linux-amd64.tar.gz\"";
    }
    next;
}
/sha256 "REPLACE_WITH_ARM64_SHA256"/ && platform == "macos" {
    print "      sha256 \"" darwin_arm64 "\"";
    next;
}
/sha256 "REPLACE_WITH_AMD64_SHA256"/ && platform == "macos" {
    print "      sha256 \"" darwin_amd64 "\"";
    next;
}
/sha256 "REPLACE_WITH_LINUX_ARM64_SHA256"/ {
    print "      sha256 \"" linux_arm64 "\"";
    next;
}
/sha256 "REPLACE_WITH_LINUX_AMD64_SHA256"/ {
    print "      sha256 \"" linux_amd64 "\"";
    next;
}
{ print; }
' > "$FORMULA_FILE.new"

mv "$FORMULA_FILE.new" "$FORMULA_FILE"
rm -f "$FORMULA_FILE.tmp" "$FORMULA_FILE.bak"

echo "✅ Homebrew formula updated successfully!"
echo ""
echo "Next steps:"
echo "1. Review the changes: git diff $FORMULA_FILE"
echo "2. Commit the changes: git add $FORMULA_FILE && git commit -m 'chore: update Homebrew formula to $VERSION'"
echo "3. Push to tap repository (if separate): git push origin main"
