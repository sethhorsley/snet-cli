# Quick Reference Card

Fast reference for common operations.

## 🚀 Release a New Version

```bash
# 1. Tag and push
git tag -a v1.0.0 -m "Release v1.0.0: Description"
git push origin v1.0.0

# 2. Watch GitHub Actions (5-10 min)
# https://github.com/sethhorsley/snet-cli/actions

# 3. Verify release
# https://github.com/sethhorsley/snet-cli/releases

# 4. Test install
brew install sethhorsley/snet-cli/snet
```

## 📦 Installation Commands

```bash
# Homebrew (direct from repo)
brew install sethhorsley/snet-cli/snet

# Homebrew (from tap - requires setup)
brew install seth4242/tap/snet

# Direct download (macOS ARM)
curl -L https://github.com/sethhorsley/snet-cli/releases/latest/download/snet-darwin-arm64.tar.gz | tar xz

# From source
go install github.com/seth4242/snet@latest
```

## 🔨 Build Commands

```bash
# Development build (localhost:3001)
make build
./bin/snet version

# Production build (seth4242.net)
make build-prod
./bin/snet version

# Install locally
make install         # Development
make install-prod    # Production

# Multi-platform release
make release

# Clean
make clean
```

## 🧪 Testing

```bash
# Run all tests
make test
go test -v ./...

# Run single test
go test -v ./internal/api -run TestName

# Test Homebrew formula
brew install --build-from-source homebrew/snet.rb
snet version
brew uninstall snet

# Test from GitHub
brew uninstall snet
brew install sethhorsley/snet-cli/snet
snet version
```

## 📝 Update Homebrew Formula

```bash
# Automatic (happens on release)
# Just push a tag - workflow handles it

# Manual
./scripts/update-homebrew-formula.sh v1.0.0
git add homebrew/snet.rb
git commit -m "chore: update Homebrew formula to v1.0.0"
git push
```

## 🔍 Verify Release

```bash
# Check GitHub Release exists
open https://github.com/sethhorsley/snet-cli/releases

# Check formula updated
git diff homebrew/snet.rb

# Verify checksums
curl -sSL https://github.com/sethhorsley/snet-cli/releases/download/v1.0.0/snet-darwin-arm64.tar.gz.sha256

# Test download
curl -L https://github.com/sethhorsley/snet-cli/releases/latest/download/snet-darwin-arm64.tar.gz | tar xz
./snet-darwin-arm64 version
```

## 📂 Key Files

```
.github/workflows/build.yml           # Main CI/CD
.github/workflows/update-homebrew.yml # Auto-update formula
.github/RELEASE_CHECKLIST.md          # Release steps
homebrew/snet.rb                      # Homebrew formula
scripts/update-homebrew-formula.sh    # Manual updater
HOMEBREW_SETUP.md                     # Complete guide
```

## 🆘 Common Issues

### Build fails
```bash
# Test locally first
go mod tidy
make build-prod
```

### Formula has wrong checksums
```bash
# Re-run update script
./scripts/update-homebrew-formula.sh v1.0.0
```

### Homebrew install fails
```bash
# User should try:
brew update
brew uninstall snet
rm -rf $(brew --cache)/snet--*
brew install sethhorsley/snet-cli/snet
```

### Workflow doesn't trigger
```bash
# Ensure tag starts with 'v'
git tag -a v1.0.0 -m "Release"
git push origin v1.0.0

# Check Actions enabled
# Repo Settings > Actions > Allow all actions
```

## 🎯 Version Numbering

```
v1.0.0  → First stable release
v1.1.0  → New feature (minor)
v1.0.1  → Bug fix (patch)
v2.0.0  → Breaking change (major)
```

## 🔗 Quick Links

- **Actions:** https://github.com/sethhorsley/snet-cli/actions
- **Releases:** https://github.com/sethhorsley/snet-cli/releases
- **Issues:** https://github.com/sethhorsley/snet-cli/issues
- **Formula Cookbook:** https://docs.brew.sh/Formula-Cookbook
- **Semantic Versioning:** https://semver.org

## 📋 Pre-Release Checklist

- [ ] Code merged to `main`
- [ ] Documentation updated
- [ ] Version number decided
- [ ] Tag created and pushed
- [ ] GitHub Actions completed
- [ ] Release verified
- [ ] Installation tested
- [ ] Announced (if applicable)

## 🎉 Success Criteria

After releasing v1.0.0, you should see:

✅ GitHub Release at `/releases/tag/v1.0.0`  
✅ 10 files in release (5 binaries + 5 checksums)  
✅ `homebrew/snet.rb` updated with v1.0.0  
✅ `brew install sethhorsley/snet-cli/snet` works  
✅ `snet version` shows v1.0.0  

## 💡 Tips

- Always test locally before releasing
- Use semantic versioning consistently
- Keep changelog updated (if you have one)
- Tag messages should describe what's new
- Monitor GitHub Actions for failures
- Test Homebrew install after each release

---

**Need more details?** See:
- Complete guide: `HOMEBREW_SETUP.md`
- Release process: `.github/RELEASE_CHECKLIST.md`
- Workflow docs: `.github/workflows/README.md`
