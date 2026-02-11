# snet

A CLI tool for creating secure HTTPS tunnels from localhost to public URLs on seth4242.net.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install seth4242/tap/snet
```

### From Source

```bash
go install github.com/seth4242/snet@latest
```

### Prerequisites

You need `cloudflared` installed:

```bash
brew install cloudflared
```

## Quick Start

```bash
# Login (one-time)
snet login

# Start a tunnel to localhost:3000
snet start --port 3000
```

This gives you a URL like `https://abc123.youraccount.seth4242.net` that proxies to your local server.

## Commands

### `snet login`

Authenticate with seth4242.net. Creates an API token at https://seth4242.net/api_tokens.

```bash
snet login
```

### `snet start`

Create and start a new tunnel.

```bash
# Basic usage
snet start --port 3000

# With wildcard subdomains (for multi-tenant apps)
snet start --port 3000 --wildcard

# Persistent tunnel (survives disconnect)
snet start --port 3000 --persistent --name my-project
```

**Flags:**
- `-p, --port` - Local port to tunnel (default: 3000)
- `-w, --wildcard` - Enable wildcard subdomains (*.slug.account.seth4242.net)
- `--persistent` - Keep tunnel after CLI disconnects
- `-n, --name` - Friendly name for the tunnel

### `snet connect`

Connect to an existing persistent tunnel.

```bash
snet connect --tunnel tun_abc123 --port 3000
```

### `snet list`

List all tunnels for your account.

```bash
snet list
```

### `snet delete`

Delete a tunnel.

```bash
snet delete tun_abc123

# Skip confirmation
snet delete tun_abc123 --force
```

### `snet logout`

Remove stored credentials.

```bash
snet logout
```

## URL Format

- **Standard:** `https://{slug}.{account}.seth4242.net`
- **Wildcard:** `https://*.{slug}.{account}.seth4242.net`

Example:
```
https://abc123.mycompany.seth4242.net
https://tenant1.abc123.mycompany.seth4242.net  (wildcard)
```

## Configuration

Config is stored in `~/.snet/config.json`:

```json
{
  "api_token": "your-token",
  "account_id": "acct_abc123",
  "api_base": "https://seth4242.net/api/v1"
}
```

## Development

```bash
# Build
make build

# Install locally
make install

# Run tests
make test

# Build for all platforms
make release
```

## License

MIT
