# GitHub Actions CI/CD

This directory contains automated build and release workflows for snet-cli.

## Workflows

### `build.yml` - Build & Release

Automated builds for beta and production releases.

#### Beta Builds (Development Mode)

**Trigger**: Push to `main` branch

**Platforms**:
- macOS (amd64, arm64)
- Linux (amd64, arm64)

**Build Mode**: Development (connects to `http://localhost:3001/api/v1`)

**Artifacts**: Available for 30 days in Actions tab
- `snet-darwin-amd64-beta`
- `snet-darwin-arm64-beta`
- `snet-linux-amd64-beta`
- `snet-linux-arm64-beta`

**Usage**:
```bash
# Push to main branch
git push origin main

# Download artifacts from GitHub Actions > Build & Release > Latest run
```

#### Production Builds (Production Mode)

**Trigger**: Push a version tag (e.g., `v1.0.0`)

**Platforms**:
- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)

**Build Mode**: Production (connects to `https://seth4242.net/api/v1`)

**Artifacts**: Released as GitHub Release with tarballs/zips

**Usage**:
```bash
# Create and push a version tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions will:
# 1. Build binaries for all platforms
# 2. Create tarballs (.tar.gz) and zips (.zip)
# 3. Create a GitHub Release
# 4. Upload all artifacts to the release
```

#### Test Builds (Pull Requests)

**Trigger**: Pull request to `main` branch

**Action**: Builds and tests code without creating artifacts

**Usage**: Automatic on PR creation/update

## Manual Trigger

All workflows can be manually triggered:

1. Go to **Actions** tab
2. Select **Build & Release** workflow
3. Click **Run workflow**
4. Choose branch

## Version Management

### Beta Versions
- Auto-generated from git: `git describe --tags --always --dirty`
- Example: `v1.0.0-5-g1234abc` (5 commits past v1.0.0, commit hash 1234abc)
- Or: `dev-1234abc` (if no tags exist)

### Production Versions
- Extracted from git tag
- Must follow semantic versioning: `v1.0.0`, `v2.1.3`, etc.
- Tag name becomes the version string

## Downloading Builds

### Beta Builds (from Actions)
1. Go to **Actions** tab
2. Click latest **Build & Release** run on `main` branch
3. Scroll to **Artifacts** section
4. Download desired platform binary

### Production Builds (from Releases)
1. Go to **Releases** page
2. Download desired platform archive
3. Extract: `tar xzf snet-darwin-arm64.tar.gz`

## Required Secrets

None required! The workflow uses the default `GITHUB_TOKEN`.

## Build Matrix

| Platform | Beta | Production | Notes |
|----------|------|------------|-------|
| macOS amd64 | ✅ | ✅ | Intel Macs |
| macOS arm64 | ✅ | ✅ | Apple Silicon (M1/M2/M3) |
| Linux amd64 | ✅ | ✅ | Standard x86_64 |
| Linux arm64 | ✅ | ✅ | ARM servers/Raspberry Pi |
| Windows amd64 | ❌ | ✅ | Production only |

## Example Release Process

```bash
# 1. Ensure code is ready on main
git checkout main
git pull origin main

# 2. Update version in CHANGELOG or documentation (optional)

# 3. Create and push tag
git tag -a v1.2.0 -m "Release v1.2.0: Add new features"
git push origin v1.2.0

# 4. Watch GitHub Actions build and release
# Visit: https://github.com/seth4242/snet-cli/actions

# 5. Release appears at: https://github.com/seth4242/snet-cli/releases
```

## Troubleshooting

### Build fails with "Go version not found"
- Go 1.25.0 might not be available in GitHub Actions yet
- Update `go-version` in workflow to current stable version (e.g., '1.23' or 'stable')

### Artifacts not appearing
- Check if workflow completed successfully
- Beta builds only create artifacts on `main` branch pushes
- Production builds require version tags starting with `v`

### Release not created
- Ensure tag starts with `v` (e.g., `v1.0.0`)
- Check workflow permissions (needs `contents: write`)
- Verify `GITHUB_TOKEN` has release permissions
