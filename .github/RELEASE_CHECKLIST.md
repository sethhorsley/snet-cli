# Release Checklist

Use this checklist when creating a new release.

## Pre-Release

- [ ] All features/fixes merged to `main`
- [ ] Tests passing (once implemented)
- [ ] Documentation updated
  - [ ] README.md reflects new features
  - [ ] CHANGELOG.md updated (if you have one)
- [ ] Version number decided (semantic versioning: MAJOR.MINOR.PATCH)

## Release Process

### 1. Create and Push Tag

```bash
# Ensure you're on main and up to date
git checkout main
git pull origin main

# Create annotated tag (replace X.Y.Z with version)
git tag -a vX.Y.Z -m "Release vX.Y.Z: Brief description"

# Push tag to trigger release workflow
git push origin vX.Y.Z
```

### 2. Monitor GitHub Actions

- [ ] Go to: https://github.com/sethhorsley/snet-cli/actions
- [ ] Watch "Build & Release" workflow
- [ ] Verify all jobs complete successfully:
  - [ ] `build-production` - Builds all platform binaries
  - [ ] `create-release` - Creates GitHub Release
  - [ ] `update-formula` - Updates Homebrew formula

**Expected Duration:** 5-10 minutes

### 3. Verify GitHub Release

- [ ] Go to: https://github.com/sethhorsley/snet-cli/releases
- [ ] Latest release should show version tag
- [ ] Check all artifacts are present:
  - [ ] `snet-darwin-amd64.tar.gz` + `.sha256`
  - [ ] `snet-darwin-arm64.tar.gz` + `.sha256`
  - [ ] `snet-linux-amd64.tar.gz` + `.sha256`
  - [ ] `snet-linux-arm64.tar.gz` + `.sha256`
  - [ ] `snet-windows-amd64.exe.zip` + `.sha256`
- [ ] Release notes auto-generated and look good

### 4. Verify Homebrew Formula

- [ ] Check `homebrew/snet.rb` was updated in this repo
- [ ] Verify version number matches release
- [ ] Verify SHA256 checksums are not placeholders

**If using separate tap repository:**
- [ ] Check `Formula/snet.rb` in `homebrew-tap` repo was updated
- [ ] Commit message mentions correct version

### 5. Test Installation

**Test Homebrew (macOS):**
```bash
# Uninstall if previously installed
brew uninstall snet

# Clear cache
rm -rf $(brew --cache)/snet--*

# Install latest
brew install sethhorsley/snet-cli/snet
# OR if using tap: brew install seth4242/tap/snet

# Verify version
snet version
# Should show: version: vX.Y.Z

# Test basic functionality
snet --help
```

**Test Direct Download (Linux):**
```bash
# Download and extract
curl -L https://github.com/sethhorsley/snet-cli/releases/latest/download/snet-linux-amd64.tar.gz | tar xz

# Check version
./snet-linux-amd64 version

# Cleanup
rm snet-linux-amd64
```

### 6. Update Tap Repository (Manual - If Not Using Auto-Update)

Only if you're NOT using automatic tap updates:

```bash
cd /path/to/homebrew-tap
cp /path/to/snet-cli/homebrew/snet.rb Formula/snet.rb
git add Formula/snet.rb
git commit -m "snet: update to vX.Y.Z"
git push origin main
```

### 7. Announce Release

- [ ] Post to social media (if applicable)
- [ ] Update project website (if applicable)
- [ ] Notify users in Discord/Slack (if applicable)
- [ ] Send email to subscribers (if applicable)

## Post-Release

- [ ] Monitor GitHub Issues for bug reports
- [ ] Check Homebrew installation reports from users
- [ ] Watch for SHA256 mismatch errors

## Rollback (If Needed)

If the release has critical bugs:

### Option 1: Quick Patch Release

```bash
# Fix the bug on main
git checkout main
# ... make fixes ...
git commit -m "fix: critical bug"
git push

# Create patch release
git tag -a vX.Y.Z+1 -m "Release vX.Y.Z+1: Fix critical bug"
git push origin vX.Y.Z+1
```

### Option 2: Delete Bad Release

```bash
# Delete tag locally and remotely
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z

# Manually delete GitHub Release in web UI
# Then create corrected release
```

## Version Numbering Guide

Following [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0 → 2.0.0): Breaking changes, incompatible API changes
- **MINOR** (1.0.0 → 1.1.0): New features, backwards compatible
- **PATCH** (1.0.0 → 1.0.1): Bug fixes, backwards compatible

### Examples

- `v1.0.0` - First stable release
- `v1.1.0` - Added new `--no-wildcard` flag
- `v1.1.1` - Fixed authentication bug
- `v2.0.0` - Changed config file format (breaking change)

### Pre-releases (Optional)

- `v1.0.0-alpha.1` - Alpha release
- `v1.0.0-beta.1` - Beta release
- `v1.0.0-rc.1` - Release candidate

## Troubleshooting

### Workflow doesn't trigger
**Problem:** Pushed tag but no workflow runs

**Solutions:**
- Ensure tag starts with `v` (e.g., `v1.0.0` not `1.0.0`)
- Check `.github/workflows/build.yml` exists in main branch
- Verify GitHub Actions are enabled in repo settings

### Build fails
**Problem:** GitHub Actions workflow fails during build

**Solutions:**
- Check workflow logs for specific error
- Common issues:
  - Go version not available (update workflow to use stable Go)
  - Import errors (run `go mod tidy` and commit)
  - Build errors (test locally: `make build-prod`)

### Homebrew formula not updated
**Problem:** Formula still has old version or placeholder SHA256s

**Solutions:**
- Check if "Update Homebrew Formula" workflow ran
- Manually run update script: `./scripts/update-homebrew-formula.sh vX.Y.Z`
- Check that SHA256 files exist in GitHub Release

### Users can't install via Homebrew
**Problem:** `brew install` fails

**Solutions:**
- Verify release completed successfully
- Check SHA256 checksums match:
  ```bash
  curl -sSL <url> | sha256sum
  ```
- Ask users to update Homebrew: `brew update`
- Ask users to clear cache: `rm -rf $(brew --cache)`

## First Release Special Steps

For your very first release (v1.0.0):

1. **Before creating tag:**
   - [ ] Ensure README has installation instructions
   - [ ] Add LICENSE file (MIT recommended)
   - [ ] Create initial GitHub Release manually or let workflow do it

2. **After release:**
   - [ ] Star your own repo (makes it more discoverable)
   - [ ] Add topic tags to repo (go, cli, tunnel, ngrok, homebrew)
   - [ ] Share on Reddit: r/golang, r/commandline
   - [ ] Submit to Show HN (Hacker News)

## Example: Complete Release

```bash
# 1. Prepare release
git checkout main
git pull origin main

# 2. Create tag
git tag -a v1.2.0 -m "Release v1.2.0: Add FRP provider support"

# 3. Push tag
git push origin v1.2.0

# 4. Wait for GitHub Actions (5-10 min)

# 5. Verify release at:
#    https://github.com/sethhorsley/snet-cli/releases/tag/v1.2.0

# 6. Test installation
brew uninstall snet
brew install sethhorsley/snet-cli/snet
snet version  # Should show: v1.2.0

# 7. Done! 🎉
```
