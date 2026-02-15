# Provider Configuration Guide

## Overview

The snet CLI supports multiple tunnel providers:
- **FRP (Fast Reverse Proxy)** - Self-hosted, full control (default)
- **Cloudflare** - Enterprise-grade edge network

## Default Provider: FRP

As of the latest version, **FRP is the default provider** for all new tunnels. This gives you:
- ✅ Full control over your tunnel infrastructure
- ✅ No external service dependencies
- ✅ Cost-effective self-hosting on Fly.io
- ✅ Direct connection to your dev server

## Configuration Methods

### 1. Config File (Recommended)

The default provider is stored in `~/.snet/config.json`:

```json
{
  "api_token": "your-api-token",
  "account_id": "acct_...",
  "api_base": "https://seth4242.net/api/v1",
  "default_provider": "frp"
}
```

**To change the default provider:**

```bash
# Method 1: Edit config file manually
vim ~/.snet/config.json
# Change "default_provider" to "cloudflare" or "frp"

# Method 2: Update via jq
jq '.default_provider = "cloudflare"' ~/.snet/config.json > /tmp/config.json
mv /tmp/config.json ~/.snet/config.json
```

### 2. Command-Line Flag (Per-Tunnel)

Override the config default for a specific tunnel:

```bash
# Use FRP (explicit)
snet start 3000 --provider frp

# Use Cloudflare (override config default)
snet start 3000 --provider cloudflare

# Use config default (omit --provider flag)
snet start 3000
```

### 3. Login Default

When you run `snet login`, the default provider is automatically set to **FRP**:

```bash
snet login

# Output includes:
# Login successful!
# Account: Your Account (slug)
# Default provider: frp
```

## Provider Selection Priority

The provider is determined in this order:

1. **Command-line flag** (`--provider`) - Highest priority
2. **Config file** (`default_provider` in `~/.snet/config.json`)
3. **Hardcoded default** - `frp` (if config doesn't specify)

### Examples

**Scenario 1: Config has `frp`, user doesn't specify flag**
```bash
# Config: "default_provider": "frp"
snet start 3000
# → Creates FRP tunnel
```

**Scenario 2: Config has `frp`, user specifies `cloudflare`**
```bash
# Config: "default_provider": "frp"
snet start 3000 --provider cloudflare
# → Creates Cloudflare tunnel (flag overrides config)
```

**Scenario 3: Config has `cloudflare`, user doesn't specify flag**
```bash
# Config: "default_provider": "cloudflare"
snet start 3000
# → Creates Cloudflare tunnel
```

**Scenario 4: No config default, user doesn't specify flag**
```bash
# Config: (no default_provider field)
snet start 3000
# → Creates FRP tunnel (hardcoded default)
```

## When to Use Each Provider

### Use FRP When:
- ✅ Developing locally and want self-hosted tunnels
- ✅ Need full control over tunnel infrastructure
- ✅ Want cost-effective tunneling
- ✅ Working on internal tools/prototypes
- ✅ Don't need Cloudflare's edge features

### Use Cloudflare When:
- ✅ Sharing demos with clients
- ✅ Need enterprise-grade edge network
- ✅ Want automatic DDoS protection
- ✅ Require Cloudflare's global CDN
- ✅ Need Zero Trust access controls

## Config File Reference

**Full config schema:**

```json
{
  "api_token": "string (required)",
  "account_id": "string (required)",
  "api_base": "string (required)",
  "default_provider": "frp | cloudflare (optional, defaults to 'frp')"
}
```

**Location:** `~/.snet/config.json`

**Permissions:** `0600` (read/write for owner only)

## Migration Guide

### From Cloudflare to FRP (Default)

If you were using Cloudflare tunnels and want to switch to FRP:

```bash
# 1. Update your config
jq '.default_provider = "frp"' ~/.snet/config.json > /tmp/config.json
mv /tmp/config.json ~/.snet/config.json

# 2. Create new tunnels (will use FRP)
snet start 3000

# 3. Old Cloudflare tunnels still work with:
snet connect --tunnel tun_xxx  # Auto-detects provider
```

### From FRP to Cloudflare

If you want to use Cloudflare as your default:

```bash
# 1. Update your config
jq '.default_provider = "cloudflare"' ~/.snet/config.json > /tmp/config.json
mv /tmp/config.json ~/.snet/config.json

# 2. Install cloudflared (if not already installed)
brew install cloudflared

# 3. Create new tunnels (will use Cloudflare)
snet start 3000
```

## Verification

**Check your current default provider:**

```bash
# Method 1: View config file
cat ~/.snet/config.json | jq .default_provider

# Method 2: Check login output
snet login
# Shows: "Default provider: frp"
```

**Verify provider for existing tunnels:**

```bash
snet list
# Shows provider column for each tunnel
```

## Troubleshooting

### Issue: "cloudflared not found"

**Cause:** Trying to create Cloudflare tunnel without cloudflared installed  
**Fix:** 
```bash
brew install cloudflared
# OR
snet start 3000 --provider frp  # Use FRP instead
```

### Issue: Config default not being used

**Cause:** Config file missing `default_provider` field  
**Fix:**
```bash
# Add default_provider to config
jq '. + {default_provider: "frp"}' ~/.snet/config.json > /tmp/config.json
mv /tmp/config.json ~/.snet/config.json
```

### Issue: Want to permanently use Cloudflare

**Fix:**
```bash
# Set Cloudflare as default in config
jq '.default_provider = "cloudflare"' ~/.snet/config.json > /tmp/config.json
mv /tmp/config.json ~/.snet/config.json

# Verify
cat ~/.snet/config.json | jq .default_provider
# Output: "cloudflare"
```

## Best Practices

1. **Set config default to your most-used provider** - Reduces typing
2. **Use explicit flags when sharing code** - Makes intent clear
3. **Document provider requirements** - In team READMEs
4. **Keep both providers available** - Easy switching for different use cases
5. **Review config after login** - Ensure default matches your preference

## Advanced: Multiple Providers

You can use both providers simultaneously:

```bash
# Terminal 1: FRP tunnel
snet start 3000 --provider frp
# → https://abc123.dev.seth4242.net

# Terminal 2: Cloudflare tunnel (different port)
snet start 8080 --provider cloudflare
# → https://def456.youraccount.seth4242.net
```

## Summary

- **Default provider:** FRP (as of this version)
- **Config location:** `~/.snet/config.json`
- **Config field:** `default_provider` (optional)
- **Valid values:** `"frp"` or `"cloudflare"`
- **Override:** Use `--provider` flag
- **Recommendation:** Keep config default as `"frp"` for local dev

---

**Quick Commands:**

```bash
# Check default
cat ~/.snet/config.json | jq .default_provider

# Set to FRP
jq '.default_provider = "frp"' ~/.snet/config.json > /tmp/c.json && mv /tmp/c.json ~/.snet/config.json

# Set to Cloudflare
jq '.default_provider = "cloudflare"' ~/.snet/config.json > /tmp/c.json && mv /tmp/c.json ~/.snet/config.json

# Use default (from config)
snet start 3000

# Override default
snet start 3000 --provider cloudflare
```
