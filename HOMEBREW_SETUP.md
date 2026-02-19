# Homebrew Distribution Setup

Complete guide for distributing snet-cli via Homebrew.

## 🎯 Quick Reference

### Current Installation Methods

```bash
# Method 1: Direct from this repository (works now)
brew install sethhorsley/snet-cli/snet

# Method 2: From custom tap (requires setup - see below)
brew install seth4242/tap/snet

# Method 3: Future goal - Homebrew core
brew install snet
```

## 📦 What's Included

1. **Homebrew Formula** (`homebrew/snet.rb`)
   - Multi-platform support (macOS Intel/ARM, Linux x86_64/ARM64)
   - Automatic binary selection based on OS/architecture
   - SHA256 verification for security

2. **GitHub Actions Workflows**
   - `build.yml` - Builds binaries and creates releases with SHA256 checksums
   - `update-homebrew.yml` - Automatically updates formula on new releases

3. **Update Script** (`scripts/update-homebrew-formula.sh`)
   - Manual formula updates when needed

## 🚀 Setup Options

### Option A: Use This Repository (Simplest - No Setup Required)

Users can install directly:

```bash
brew install sethhorsley/snet-cli/snet
```

**Pros:**
- ✅ No additional repository needed
- ✅ Works immediately
- ✅ Automatic updates via GitHub Actions

**Cons:**
- ❌ Longer install command
- ❌ Less discoverable

### Option B: Create Separate Tap Repository (Recommended)

Provides cleaner user experience: `brew install seth4242/tap/snet`

#### 1. Create Tap Repository

```bash
# On GitHub: Create new public repository named "homebrew-tap"
# Note: Must be named "homebrew-<something>" for Homebrew to find it

# Clone locally
git clone git@github.com:seth4242/homebrew-tap.git
cd homebrew-tap

# Create directory structure
mkdir -p Formula

# Add README
cat > README.md << 'EOF'
# Seth4242 Homebrew Tap

Official Homebrew tap for seth4242 tools.

## Installation

```bash
brew tap seth4242/tap
brew install snet
```

Or in one command:

```bash
brew install seth4242/tap/snet
```

## Available Formulae

- **snet** - Secure tunneling CLI (ngrok alternative)
EOF

git add README.md
git commit -m "Initial commit"
git push origin main
```

#### 2. Enable Automatic Updates (Optional)

To automatically update the tap on releases:

**A. Create GitHub Personal Access Token**

1. Go to: https://github.com/settings/tokens/new
2. Name: "Homebrew Tap Updates"
3. Scopes: Select `repo` (all)
4. Click "Generate token"
5. Copy the token (you won't see it again!)

**B. Add Secret to snet-cli Repository**

1. Go to: https://github.com/sethhorsley/snet-cli/settings/secrets/actions
2. Click "New repository secret"
3. Name: `TAP_REPO_TOKEN`
4. Value: Paste the token
5. Click "Add secret"

**C. Uncomment Workflow Section**

Edit `.github/workflows/update-homebrew.yml` and uncomment the section at the bottom:

```yaml
- name: Checkout homebrew-tap
  uses: actions/checkout@v4
  with:
    repository: seth4242/homebrew-tap
    token: ${{ secrets.TAP_REPO_TOKEN }}
    path: tap

- name: Update tap repository
  run: |
    cp homebrew/snet.rb tap/Formula/snet.rb
    cd tap
    git config user.name "github-actions[bot]"
    git config user.email "github-actions[bot]@users.noreply.github.com"
    git add Formula/snet.rb
    git commit -m "snet: update to ${{ steps.version.outputs.version }}"
    git push
```

Now releases automatically update both repositories!

#### 3. Manual Updates (If Not Using Automatic)

After each release:

```bash
# Update snet-cli repo
cd /path/to/snet-cli
./scripts/update-homebrew-formula.sh v1.0.0
git add homebrew/snet.rb
git commit -m "chore: update Homebrew formula to v1.0.0"
git push

# Update tap repo
cd /path/to/homebrew-tap
cp /path/to/snet-cli/homebrew/snet.rb Formula/snet.rb
git add Formula/snet.rb
git commit -m "snet: update to v1.0.0"
git push
```

## 📝 Release Process

### With Automatic Updates (Recommended)

```bash
# 1. Ensure code is ready
git checkout main
git pull

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0: Description"
git push origin v1.0.0

# 3. Wait for GitHub Actions (5-10 minutes)
#    - Builds binaries for all platforms
#    - Creates GitHub Release
#    - Generates SHA256 checksums
#    - Updates homebrew/snet.rb in this repo
#    - Updates Formula/snet.rb in tap repo (if configured)

# 4. Done! Users can now install:
#    brew upgrade snet
```

### Without Automatic Updates

```bash
# 1-3. Same as above

# 4. Manually update formula
./scripts/update-homebrew-formula.sh v1.0.0

# 5. Commit and push
git add homebrew/snet.rb
git commit -m "chore: update Homebrew formula to v1.0.0"
git push

# 6. Update tap (if separate)
cd /path/to/homebrew-tap
cp /path/to/snet-cli/homebrew/snet.rb Formula/snet.rb
git add Formula/snet.rb
git commit -m "snet: update to v1.0.0"
git push
```

## 🧪 Testing

### Test Formula Locally

```bash
# Install from local file
brew install --build-from-source homebrew/snet.rb

# Verify it works
snet version

# Uninstall
brew uninstall snet
```

### Test From Repository

```bash
# Uninstall if already installed
brew uninstall snet

# Install from GitHub
brew install sethhorsley/snet-cli/snet

# Test
snet version
```

### Test From Tap (if using separate repo)

```bash
# Add tap
brew tap seth4242/tap

# Install
brew install snet

# Test
snet version

# Remove tap (optional)
brew untap seth4242/tap
```

### Run Homebrew Audit

```bash
# Check for formula issues
brew audit --strict homebrew/snet.rb

# Test installation and tests
brew install --build-from-source homebrew/snet.rb
brew test snet
```

## 🎯 Future: Submit to Homebrew Core

Once snet-cli is stable and popular, submit to [homebrew-core](https://github.com/Homebrew/homebrew-core).

### Requirements

- ✅ Stable, versioned releases
- ✅ 75+ GitHub stars (recommended)
- ✅ Good documentation
- ✅ Active maintenance
- ✅ Open source license
- ✅ No major security issues

### Process

1. Fork https://github.com/Homebrew/homebrew-core
2. Create `Formula/snet.rb` (copy from `homebrew/snet.rb`)
3. Submit PR with clear description
4. Address reviewer feedback
5. Once merged: users can `brew install snet`

See: https://docs.brew.sh/Adding-Software-to-Homebrew

## 🔧 Troubleshooting

### "Error: No available formula with the name"

**Problem:** Homebrew can't find the formula

**Solutions:**
- Repository name must be `homebrew-<name>` (e.g., `homebrew-tap`)
- Formula must be in `Formula/` directory
- Repository must be public
- Try: `brew tap --force seth4242/tap`

### "Error: SHA256 mismatch"

**Problem:** Downloaded file doesn't match expected checksum

**Solutions:**
- Re-run update script: `./scripts/update-homebrew-formula.sh v1.0.0`
- Verify checksums in GitHub release
- Ensure version tag matches release

### Formula Update Workflow Fails

**Problem:** GitHub Actions can't update tap repository

**Solutions:**
- Verify `TAP_REPO_TOKEN` secret exists
- Check token has `repo` scope
- Ensure repository name is correct in workflow
- Check workflow logs for specific error

### Users Get Old Version

**Problem:** `brew install snet` installs outdated version

**Solutions:**
- Update formula: `./scripts/update-homebrew-formula.sh vX.Y.Z`
- Users run: `brew update && brew upgrade snet`
- Check tap repository has latest formula

## 📚 Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [How to Create Homebrew Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Formula Documentation](https://rubydoc.brew.sh/Formula)

## 🎉 Current Status

- ✅ Homebrew formula created
- ✅ GitHub Actions builds releases
- ✅ SHA256 checksums generated
- ✅ Auto-update workflow configured
- ✅ Update script available
- ⏳ Separate tap repository (optional - needs setup)
- ⏳ Homebrew core submission (future goal)
