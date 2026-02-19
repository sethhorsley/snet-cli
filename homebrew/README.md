# Homebrew Distribution

This directory contains the Homebrew formula for installing snet-cli.

## Two Installation Methods

### 1. Custom Tap (Current Method)

Users can install from your custom tap:

```bash
brew install seth4242/tap/snet
```

This requires a separate repository at `github.com/seth4242/homebrew-tap` (see setup below).

### 2. Homebrew Core (Future Goal)

Eventually you can submit to Homebrew core, allowing:

```bash
brew install snet
```

This requires submitting a PR to [Homebrew/homebrew-core](https://github.com/Homebrew/homebrew-core).

## Setting Up Your Homebrew Tap

### Option A: Using This Repository (Simplest)

If you want to keep the formula in this repo, users can install with:

```bash
brew install sethhorsley/snet-cli/snet
```

**No additional setup needed!** Homebrew can read formulas from any GitHub repo.

### Option B: Separate Tap Repository (Recommended)

For a cleaner experience (`brew install seth4242/tap/snet`), create a separate tap:

#### 1. Create Tap Repository

```bash
# Create new repository on GitHub: homebrew-tap
# Clone it locally
git clone git@github.com:seth4242/homebrew-tap.git
cd homebrew-tap

# Copy formula
mkdir -p Formula
cp /path/to/snet-cli/homebrew/snet.rb Formula/snet.rb

# Commit and push
git add Formula/snet.rb
git commit -m "Add snet formula"
git push origin main
```

#### 2. Users Install From Tap

```bash
brew tap seth4242/tap
brew install snet
```

Or in one command:
```bash
brew install seth4242/tap/snet
```

## Updating the Formula

### After Each Release

When you release a new version (e.g., `v1.0.0`):

#### 1. Wait for GitHub Actions to Complete

The CI workflow will:
- Build binaries for all platforms
- Create `.tar.gz` archives
- Generate `.sha256` checksum files
- Create a GitHub Release

#### 2. Update Formula with Script

```bash
# Run the update script
./scripts/update-homebrew-formula.sh v1.0.0

# Review changes
git diff homebrew/snet.rb

# Commit changes
git add homebrew/snet.rb
git commit -m "chore: update Homebrew formula to v1.0.0"
git push
```

#### 3. Update Tap Repository (if separate)

```bash
cd /path/to/homebrew-tap
cp /path/to/snet-cli/homebrew/snet.rb Formula/snet.rb
git add Formula/snet.rb
git commit -m "snet: update to v1.0.0"
git push origin main
```

### Manual Updates

If you prefer manual updates, edit `homebrew/snet.rb`:

1. Update version number:
   ```ruby
   version "1.0.0"
   ```

2. Update URLs:
   ```ruby
   url "https://github.com/sethhorsley/snet-cli/releases/download/v1.0.0/snet-darwin-arm64.tar.gz"
   ```

3. Update SHA256 checksums:
   ```bash
   # Download release and compute SHA256
   curl -sSL https://github.com/sethhorsley/snet-cli/releases/download/v1.0.0/snet-darwin-arm64.tar.gz | sha256sum
   ```

   ```ruby
   sha256 "abc123..."
   ```

## Testing the Formula

### Test Installation Locally

```bash
# Install from local file
brew install --build-from-source homebrew/snet.rb

# Test the binary
snet version

# Uninstall
brew uninstall snet
```

### Test from Tap

```bash
# If using this repo
brew install sethhorsley/snet-cli/snet

# If using separate tap
brew tap seth4242/tap
brew install snet
```

### Run Formula Audit

```bash
# Check formula for issues
brew audit --strict homebrew/snet.rb

# Test installation in clean environment
brew install --build-from-source homebrew/snet.rb
brew test snet
```

## Submitting to Homebrew Core

Once your CLI is stable and popular, submit to Homebrew core:

### Requirements

1. **Notable project** - Should have substantial usage/stars
2. **Stable releases** - Regular release cycle with proper versioning
3. **Good documentation** - Clear README and usage docs
4. **Active maintenance** - Responsive to issues
5. **License** - Open source license (MIT, Apache, etc.)

### Submission Process

1. Fork [Homebrew/homebrew-core](https://github.com/Homebrew/homebrew-core)

2. Create formula in `Formula/snet.rb`

3. Submit PR with:
   - Formula file
   - Clear description
   - Link to project
   - Why it should be in core

4. Address reviewer feedback

5. Once merged, users can: `brew install snet`

See: https://docs.brew.sh/Adding-Software-to-Homebrew

## Formula Structure

The formula supports:
- ✅ macOS Intel (x86_64)
- ✅ macOS Apple Silicon (ARM64)
- ✅ Linux Intel (x86_64)
- ✅ Linux ARM64

It automatically selects the correct binary based on:
- OS: `on_macos` / `on_linux`
- Architecture: `Hardware::CPU.arm?` / `Hardware::CPU.intel?`

## Current Version

The formula in this repository is a **template**. Run the update script after creating your first release to populate with real checksums.

## Troubleshooting

### Formula fails audit
```bash
brew audit --strict homebrew/snet.rb
```

Common issues:
- Missing license
- Incorrect URL format
- Invalid SHA256
- Missing test block

### Installation fails
- Check binary is executable in tarball: `tar -tzf snet-darwin-arm64.tar.gz`
- Verify SHA256: `sha256sum snet-darwin-arm64.tar.gz`
- Test download: `curl -sSL <url> | tar -xz`

### Users can't find tap
- Ensure repository is public
- Check naming: must be `homebrew-<tap-name>`
- Verify formula is in `Formula/` directory
