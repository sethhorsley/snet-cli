# Quick Start: Using snet with FRP Tunnels

## Current Setup Status

✅ **FRP Server**: Running on Fly.io (149.248.211.110)  
✅ **snet CLI**: Built with FRP support  
✅ **Rails API**: Migrated with FRP fields  
✅ **Test Tunnel**: Currently running (PID 52805)

## Using snet CLI with FRP

### 1. Start an FRP Tunnel

```bash
cd /Users/send16/files/sethhorsley/snet-cli

# Create a new FRP tunnel
./bin/snet start 3000 --provider frp --name "my-dev-tunnel"
```

**Output:**
```
Creating frp tunnel...
Tunnel created successfully!

⚠️  FRP tunnels require frpc to be installed.

To install frpc:
  brew install frpc

Your tunnel is configured! To connect manually:

  frpc tcp \
    --server_addr 149.248.211.110 \
    --server_port 7000 \
    --token LqwuyC9iYRtWEEEawHdHOd3YtiSBt9ffyyykng58hXU= \
    --proxy_name testuser-abc12345 \
    --type http \
    --local_ip localhost \
    --local_port 3000 \
    --subdomain abc12345

Tunnel URL: https://abc12345.testuser.dev.seth4242.net

Press Ctrl+C when done...
```

### 2. Run the FRP Client

**Option A: Install frpc**
```bash
brew install frpc
```

**Option B: Use existing frpc**  
The command is already shown in the snet output. Copy and run it in a new terminal.

### 3. Stop a Tunnel

```bash
# Stop all tunnels
./bin/snet stop

# Stop specific tunnel by slug
./bin/snet stop abc12345

# Stop by process ID
./bin/snet stop 52805
```

### 4. List All Tunnels

```bash
./bin/snet list
```

## Manual Tunnel (Current Method)

Until you restart Rails and test the new snet CLI, you can still use the manual method:

```bash
# The tunnel is already running!
# Access it at: https://test.dev.seth4242.net

# To stop it:
kill 52805  # Or: kill $(cat frpc.pid)

# To restart it:
./start-frp-tunnel.sh
```

## Next Steps to Test Full Integration

### 1. Restart Rails Server

```bash
cd /Users/send16/files/sethhorsley/seth4242.net

# Stop current server (Ctrl+C in the terminal where it's running)
# Then restart:
bin/dev
```

### 2. Test snet CLI

```bash
cd /Users/send16/files/sethhorsley/snet-cli

# Create a new tunnel
./bin/snet start 3000 --provider frp --name "snet-test"

# This should:
# ✅ Create tunnel in database
# ✅ Return FRP credentials
# ✅ Show frpc command to run
# ✅ Start sending heartbeats
```

### 3. Run the FRP Client

Copy the `frpc tcp ...` command from snet output and run it:

```bash
frpc tcp \
  --server_addr 149.248.211.110 \
  --server_port 7000 \
  --token <shown-in-output> \
  --proxy_name <shown-in-output> \
  --type http \
  --local_ip localhost \
  --local_port 3000 \
  --subdomain <shown-in-output>
```

### 4. Test the Tunnel

```bash
# Get the URL from snet output, e.g.:
curl https://<slug>.<account>.dev.seth4242.net/up

# Should return green HTML page (200 OK)
```

## Troubleshooting

### Error: "No route matches [POST] /api/v1/tunnels"

**Cause**: Rails server needs restart to load new code  
**Fix**: Restart Rails with `bin/dev`

### Error: "frpc: command not found"

**Cause**: frpc not installed  
**Fix**: Run `brew install frpc`

### Error: "failed to create tunnel: 401 Unauthorized"

**Cause**: Invalid API token  
**Fix**: Check `~/.snet/config.json` has correct token and account_id

### Tunnel created but not accessible

**Checks**:
1. Is frpc process running? `ps aux | grep frpc`
2. Is Rails server on localhost:3000? `curl localhost:3000/up`
3. Is FRP server reachable? `nc -zv 149.248.211.110 7000`
4. Check FRP server logs: `cd fly-frp && fly logs -a seth4242-frp`

## Architecture Summary

```
User Runs:
  snet start 3000 --provider frp
      ↓
snet CLI:
  1. Calls Rails API to create tunnel
  2. Receives FRP credentials (auth token + proxy name)
  3. Displays frpc command for user
  4. Sends periodic heartbeats to API
      ↓
User Runs (manually):
  frpc tcp --server_addr 149.248.211.110 ...
      ↓
FRP Client:
  1. Connects to FRP server on Fly.io
  2. Registers proxy with subdomain
  3. Proxies traffic: Internet → FRP Server → Local Port 3000
      ↓
Result:
  https://<slug>.<account>.dev.seth4242.net → localhost:3000
```

## Current Limitations

1. **Manual frpc Execution**: User must run frpc command separately
2. **No Auto-Retry**: If frpc crashes, user must restart manually
3. **No Process Management**: snet doesn't manage frpc lifecycle

## Future Improvements

1. **Embed frpc**: Bundle binary in snet or download on first run
2. **Auto-Start**: snet spawns frpc subprocess automatically
3. **Process Monitoring**: Detect and restart crashed frpc
4. **Better UX**: Progress bars, auto-copy commands, status indicators

## Files Changed

### snet-cli
- `cmd/start.go` - Added --provider flag, FRP routing
- `cmd/stop.go` - New stop command
- `cmd/connect.go` - Auto-detect provider
- `internal/api/client.go` - FRP fields in Tunnel struct
- `internal/tunnel/frp_runner.go` - FRP tunnel runner
- `go.mod` - Added FRP dependencies

### seth4242.net (Rails)
- `db/migrate/..._add_frp_support_to_tunnels.rb` - New fields
- `app/models/tunnel.rb` - Provider enum, FRP encryption
- `app/models/tunnel/provisioning.rb` - FRP provisioning
- `app/controllers/api/v1/tunnels_controller.rb` - FRP JSON fields
- `.env.development.local` - FRP_AUTH_TOKEN

---

**Current Status**: ✅ All code complete, ready for testing after Rails restart

**Test Tunnel**: https://test.dev.seth4242.net (currently accessible)
