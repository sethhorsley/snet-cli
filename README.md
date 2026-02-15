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

### No External Dependencies

snet is a single binary with **no external dependencies**. The FRP (Fast Reverse Proxy) client is embedded directly into the binary, so you don't need to install `cloudflared`, `frpc`, or any other tools.

## Quick Start

```bash
# Login (one-time)
snet login

# Start a tunnel to localhost:3000
snet start 3000

# Press Ctrl+C to stop - automatically cleans up tunnel and DNS records
```

This gives you a URL like `https://abc123.youraccount.seth4242.net` that proxies to your local server.

**When you stop the tunnel (Ctrl+C), it automatically removes all DNS records and Cloudflare resources for ephemeral tunnels.**

## Global Flags

These flags work with any command:

| Flag | Description |
|------|-------------|
| `-q, --quiet` | Suppress non-essential output |
| `-v, --verbose` | Enable verbose output |
| `--api-port` | Override API port (development mode only) |

## Commands

### `snet login`

Authenticate with seth4242.net. Creates an API token at https://seth4242.net/api_tokens.

```bash
snet login
```

### `snet start [port]`

Create and start a new tunnel.

**Wildcard subdomains are ENABLED BY DEFAULT** - every tunnel supports `*.tunnel.account.seth4242.net`

**Press Ctrl+C to stop** - automatically cleans up DNS records and removes the tunnel (unless `--persistent`)

```bash
# Basic usage (defaults to port 3000 with wildcards)
snet start
snet start 3000
snet start 8080

# Disable wildcard subdomains
snet start 3000 --no-wildcard

# Request a custom subdomain
snet start 3000 --subdomain myapp   # https://myapp.account.seth4242.net

# Tunnel to a different host (not localhost)
snet start 8080 --host 192.168.1.100

# Persistent tunnel (survives disconnect, NOT auto-cleaned)
snet start 3000 --persistent --name my-project

# Quiet mode (only outputs the URL)
snet start 3000 -q
```

**Flags:**
| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--port` | `-p` | 3000 | Local port to tunnel (can also use positional arg) |
| `--host` | | localhost | Host to tunnel to |
| `--subdomain` | `-s` | | Request a specific subdomain |
| `--wildcard` | `-w` | **true** | Enable wildcard subdomains (default) |
| `--no-wildcard` | | false | Disable wildcard subdomains |
| `--persistent` | | false | Keep tunnel after CLI disconnects |
| `--name` | `-n` | | Friendly name for the tunnel |

### `snet connect [port]`

Connect to an existing persistent tunnel.

```bash
snet connect --tunnel tun_abc123           # port 3000
snet connect --tunnel tun_abc123 8080      # port 8080
snet connect -t tun_abc123 --host 10.0.0.5 # tunnel to remote host
```

**Flags:**
| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--tunnel` | `-t` | | Tunnel ID to connect to (required) |
| `--port` | `-p` | 3000 | Local port (can also use positional arg) |
| `--host` | | localhost | Host to tunnel to |

### `snet list`

List all tunnels for your account.

```bash
snet list             # table format
snet ls               # alias

# JSON output for scripting
snet list --json
snet list --json | jq '.[] | select(.status=="active")'
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |

### `snet delete`

Delete a tunnel.

```bash
snet delete tun_abc123
snet rm tun_abc123        # alias

# Skip confirmation
snet delete tun_abc123 -y
snet delete tun_abc123 --force
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--yes` | `-y` | Skip confirmation |
| `--force` | `-f` | Skip confirmation (alias) |

### `snet logout`

Remove stored credentials.

```bash
snet logout
```

### `snet version`

Show version and build info.

```bash
snet version
```

## URL Format

- **Standard:** `https://{slug}.{account}.seth4242.net`
- **Wildcard:** `https://*.{slug}.{account}.seth4242.net`

Example:
```
https://abc123.mycompany.seth4242.net
https://tenant1.abc123.mycompany.seth4242.net  (wildcard)
```

## Automatic HTTPS & SSL

Every tunnel automatically gets:
- **DNS Records**: Cloudflare DNS automatically configured
- **SSL Certificates**: Valid HTTPS certificates via Fly.io (Let's Encrypt)
- **ACME Validation**: DNS-01 challenge records automatically created
- **Zero Configuration**: Everything happens when you run `snet start`

The SSL certificates typically validate within 30-60 seconds. Your tunnel works immediately over HTTP, and switches to HTTPS once certificates are validated.

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

### Build Modes

The CLI has two build modes that determine which API endpoint it connects to:

| Mode | API Endpoint | Use Case |
|------|--------------|----------|
| **development** | `http://localhost:3001/api/v1` | Local Rails server |
| **production** | `https://seth4242.net/api/v1` | Live service |

Check current mode with `snet version`:

```
snet dev
  mode: development
  api: http://localhost:3001/api/v1
  go: go1.21.0
  os/arch: darwin/arm64
```

### Development Builds (default)

Development mode connects to `localhost:3001` for testing against a local Rails server:

```bash
cd snet-cli

# Fastest iteration - run without building (development mode)
go run . start --port 3000

# Build then run (development mode)
make build && ./bin/snet start --port 3000

# Install globally (development mode)
make install
```

### Production Builds

Production mode connects to `seth4242.net`:

```bash
# Build production binary
make build-prod && ./bin/snet version

# Install production binary globally
make install-prod

# Build release binaries for all platforms (always production)
make release
```

### Make Commands

```bash
# Development mode (localhost:3001)
make build      # Build to bin/snet
make install    # Install to $GOPATH/bin

# Production mode (seth4242.net)
make build-prod # Build to bin/snet
make install-prod # Install to $GOPATH/bin

# Other
make test       # Run tests
make release    # Build for all platforms (production)
make clean      # Remove build artifacts
```

## License

MIT
