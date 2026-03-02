# Troubleshooting Guide

## Understanding Tunnel Creation Messages

### Normal Flow

When creating a new tunnel, you'll see:

```
Creating frp tunnel...
  → Provisioning wildcard support
  → Generating authentication credentials
✓ Tunnel created successfully!
```

Here's what each step means:

1. **"Creating frp tunnel..."** - Starting tunnel creation process
2. **"Provisioning wildcard support"** - Configuring DNS for `*.yourtunnel.account.snet-public.com`
3. **"Generating authentication credentials"** - The Rails API is:
   - Reading the shared FRP auth token from Rails credentials
   - Generating a unique proxy name for this tunnel
   - Creating DNS records in Cloudflare
   - Requesting SSL certificates from Fly.io

### Common Errors

#### Authentication Error (Code 10000)

```
failed to create tunnel: API error (422): Failed to provision FRP tunnel: 
{"success":false,"errors":[{"code":10000,"message":"Authentication error"}]}
```

**What this means:**
- This error happens on the **Rails server**, not in the CLI
- The Rails server is trying to create DNS records in Cloudflare
- The Cloudflare API token stored in Rails credentials is **invalid or expired**

**The Rails tunnel provisioning flow:**
1. ✅ Read FRP auth token from `Rails.application.credentials.dig(:frp, :auth_token)`
2. ✅ Generate proxy name: `{account-slug}-{tunnel-slug}`
3. ❌ Create DNS records in Cloudflare → **FAILS HERE with Authentication error**
4. ⏭️  Request SSL certificates (skipped due to step 3 failure)

**How to fix:**
1. Go to Cloudflare Dashboard → API Tokens
2. Create a new token with these permissions:
   - **Zone.DNS** - Edit
   - **Zone.Zone** - Read
   - For zone: `snet-public.com`
3. Update Rails credentials:
   ```bash
   cd /path/to/seth4242.net
   bin/rails credentials:edit
   ```
4. Update the token:
   ```yaml
   cloudflare:
     account_id: abe265a4b62b72738164d20c967bb3b6
     zone_id: c3a212d5b0191a2e25f0f5b768d2a883
     api_token: YOUR_NEW_TOKEN_HERE
   ```
5. Save and try creating the tunnel again

#### FRP Token Mismatch

```
token in login doesn't match token from configuration
```

**What this means:**
- The FRP server (running on Fly.io) has a different token than what Rails provided
- This happens when the FRP server is restarted with a different `FRP_AUTH_TOKEN` secret

**How to fix:**
The CLI now **automatically handles this** by deleting and recreating the tunnel with fresh credentials. If automatic recovery fails, manually sync the tokens:

1. Check Rails credentials token:
   ```bash
   cd /path/to/seth4242.net
   bin/rails credentials:show | grep -A2 "frp:"
   ```
2. Update Fly.io secret to match:
   ```bash
   cd /path/to/seth4242.net/fly-frp
   fly secrets set FRP_AUTH_TOKEN="token-from-rails"
   ```

## Verbose Mode

Use `-v` or `--verbose` to see detailed debugging information:

```bash
snet http 3000 -v
```

### Creating a New Tunnel (Verbose)

```
[DEBUG] Creating tunnel with parameters:
  Name:       my-app
  Port:       3000
  Wildcard:   true
  Persistent: true
  Provider:   frp
  API:        https://seth4242.net/api/v1

[DEBUG] Tunnel creation failed:
  Error: API error (422): ...

[EXPLANATION]
This error occurs on the Rails API server during tunnel provisioning.

The Rails server follows these steps:
  1. Generate FRP authentication token (from Rails credentials)
  2. Create DNS records in Cloudflare (*.my-app.ACCOUNT.snet-public.com)
  3. Request SSL certificates from Fly.io

The 'Authentication error' indicates step 2 failed.
...
```

### Reconnecting to Existing Tunnel (Verbose)

```
[DEBUG] Reconnecting to existing tunnel:
  ID:              tun_abc123
  Name:            my-app
  URL:             https://abc123.seth.snet-public.com
  Provider:        frp
  FRP Server:      snet-frp.fly.dev:7000
  FRP Proxy Name:  seth-abc123
  FRP Auth Token:  Lqwu...hXU=
```

This shows:
- Which tunnel you're connecting to
- The FRP server address and port
- The authentication token being used (masked for security)

## Architecture Overview

### FRP Tunnel Creation Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ CLI: snet http 3000                                             │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   │ POST /api/v1/tunnels
                   │ {name, port, wildcard, provider: "frp"}
                   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Rails API: Tunnel.provision_frp!                                │
├─────────────────────────────────────────────────────────────────┤
│ 1. Read FRP_AUTH_TOKEN from credentials ✅                      │
│ 2. Generate proxy name: {account}-{slug} ✅                     │
│ 3. Create DNS records in Cloudflare ❌ (AUTH ERROR)            │
│ 4. Request SSL certificates (skipped)                           │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   │ ERROR 422: Authentication error
                   ↓
┌─────────────────────────────────────────────────────────────────┐
│ CLI: Displays error                                             │
│      (with verbose: shows detailed explanation)                 │
└─────────────────────────────────────────────────────────────────┘
```

### Token Architecture

The system uses **two different authentication mechanisms**:

1. **FRP Authentication** (Tunnel connections)
   - **Single shared token** used by all tunnels
   - Stored in Rails: `credentials.dig(:frp, :auth_token)`
   - Stored in Fly: `FRP_AUTH_TOKEN` secret
   - Used by: FRP clients → FRP server on Fly.io

2. **Cloudflare Authentication** (DNS management)
   - **API token** for creating/deleting DNS records
   - Stored in Rails: `credentials.dig(:cloudflare, :api_token)`
   - Used by: Rails API → Cloudflare DNS API
   - Required permissions: Zone.DNS (Edit), Zone.Zone (Read)

3. **Fly.io Authentication** (SSL certificates)
   - **API token** for requesting SSL certificates
   - Stored in Rails: `credentials.dig(:fly, :api_token)`
   - Used by: Rails API → Fly.io GraphQL API

## Quick Reference

### Check Current Configuration

```bash
# Check Rails credentials
cd /path/to/seth4242.net
bin/rails credentials:show

# Check Fly secrets
cd /path/to/seth4242.net/fly-frp
fly secrets list

# Test Cloudflare token
curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

### Fix Authentication Issues

```bash
# 1. Update Cloudflare token in Rails
cd /path/to/seth4242.net
bin/rails credentials:edit
# Update: cloudflare.api_token

# 2. Sync FRP token to Fly.io
cd /path/to/seth4242.net/fly-frp
fly secrets set FRP_AUTH_TOKEN="$(bin/rails credentials:show | grep 'auth_token:' | awk '{print $2}')"

# 3. Test tunnel creation
cd /path/to/snet-cli
./snet http 3000 -v
```
