# Header Configuration in snet-cli

Framework-agnostic header customization for FRP tunnels.

> **⚠️ Important:** Always quote header values containing special characters like `*`, `$`, `&`:
> ```bash
> --response-header "Access-Control-Allow-Origin:*"  # ✅ Correct
> --response-header Access-Control-Allow-Origin:*    # ❌ Shell error!
> ```

## Default Behavior: Transparent Mode

By default, snet-cli operates in **transparent mode** - all headers pass through unchanged:

```bash
snet start 3000
# Headers pass through as-is
# Host: tunnel.account.seth4242.net (original)
# X-Forwarded-For: <client-ip> (auto-added by FRP)
# X-Forwarded-Proto: https (auto-added by FRP)
```

## Custom Headers

### Add Request Headers

Add custom headers to requests sent to your local server:

```bash
snet start 3000 -H X-Api-Key:secret-token
snet start 3000 --header X-Api-Key:secret-token
```

Multiple headers:

```bash
snet start 3000 \
  -H X-Api-Key:secret-token \
  -H X-Environment:development \
  -H X-Debug-Mode:true
```

### Add Response Headers

Add custom headers to responses sent to clients:

```bash
snet start 8080 --response-header "Access-Control-Allow-Origin:*"
```

**Note:** Quote values containing `*` or other shell special characters.

Multiple response headers:

```bash
snet start 8080 \
  --response-header "Access-Control-Allow-Origin:*" \
  --response-header "Access-Control-Allow-Methods:GET,POST,PUT,DELETE" \
  --response-header "X-Powered-By:snet-tunnels"
```

### Rewrite Host Header

Change the Host header to match your local development setup:

```bash
snet start 3000 --host-header localhost:3000
```

This is useful for frameworks that validate the Host header (like Rails):

```bash
# Rails development with host validation
snet start 3000 --host-header localhost:3000
# Your app sees: Host: localhost:3000
# Original host preserved in: X-Forwarded-Host
```

## Common Use Cases

### 1. Rails/Django Development

Prevent "Blocked host" errors:

```bash
snet start 3000 \
  --host-header localhost:3000 \
  -H X-Forwarded-Host:tunnel.account.seth4242.net
```

### 2. API with Custom Headers

```bash
snet start 8080 \
  -H X-Api-Version:v2 \
  -H X-Service-Name:my-api
```

### 3. CORS Configuration

```bash
snet start 8080 \
  --response-header "Access-Control-Allow-Origin:*" \
  --response-header "Access-Control-Allow-Methods:GET,POST,PUT,DELETE,OPTIONS" \
  --response-header "Access-Control-Allow-Headers:Content-Type,Authorization"
```

### 4. Custom Local Domain

If using .test or .local domains:

```bash
snet start 3000 --host-header myapp.test
```

### 5. Debugging Headers

```bash
snet start 3000 \
  -H X-Debug:true \
  -H X-Trace-ID:test-123 \
  --response-header X-Server:dev-tunnel
```

## Header Format

Headers use the format: `name:value`

- Header name and value can contain spaces (they're auto-trimmed)
- Colons in values are preserved
- **Always quote values with shell special characters** (`*`, `$`, `&`, etc.)

Example formats:
```bash
-H "X-Custom: value with spaces"
-H "Authorization: Bearer token123"
-H X-API-Key:no-spaces-needed
--response-header "Access-Control-Allow-Origin:*"  # Quote the asterisk!
```

## Combining with Other Flags

All header flags work with other snet-cli options:

```bash
# Persistent tunnel with custom headers
snet start 3000 \
  --persistent \
  --name my-api \
  -H X-Environment:staging

# Connect to existing tunnel with headers
snet connect my-api 8080 \
  --host-header localhost:8080 \
  -H X-Debug:true

# Custom subdomain with CORS
snet start 8080 \
  --subdomain api \
  --response-header "Access-Control-Allow-Origin:*"
```

## What Headers Are Automatically Added?

FRP automatically adds these headers (you cannot disable them):

```
X-Forwarded-For: <client-ip>      # Real client IP address
X-Forwarded-Proto: https           # Protocol (http or https)
X-Real-IP: <client-ip>            # Same as X-Forwarded-For
```

When you use `--host-header`, FRP also preserves:

```
X-Forwarded-Host: <original-host>  # Original tunnel hostname
```

## Testing Headers

### View Headers Your App Receives

In your application, log the headers:

**Rails:**
```ruby
# In a controller
Rails.logger.info request.headers.to_h
```

**Express.js:**
```javascript
app.get('/', (req, res) => {
  console.log('Headers:', req.headers);
  res.json({ headers: req.headers });
});
```

**Python/Flask:**
```python
from flask import request

@app.route('/')
def index():
    print('Headers:', dict(request.headers))
    return dict(request.headers)
```

### Using curl

Test from outside your network:

```bash
# See all headers
curl -v https://tunnel.account.seth4242.net

# Check specific response headers
curl -I https://tunnel.account.seth4242.net

# Send custom request headers
curl -H "X-Test: value" https://tunnel.account.seth4242.net
```

## Examples by Framework

### Next.js / React

Usually works fine in transparent mode:

```bash
snet start 3000
# No header configuration needed
```

### Ruby on Rails (Development)

Prevent host blocking:

```bash
snet start 3000 --host-header localhost:3000
```

### Django

Same as Rails:

```bash
snet start 8000 --host-header localhost:8000
```

### Express.js / Node.js

Transparent mode typically works:

```bash
snet start 3000
# Or add custom headers as needed
```

### Laravel

May need host rewrite:

```bash
snet start 8000 --host-header localhost:8000
```

### ASP.NET / .NET

Usually transparent:

```bash
snet start 5000
```

### Flask / FastAPI

Transparent mode:

```bash
snet start 8000
```

## Advanced: Custom Authentication

If your app expects specific auth headers:

```bash
snet start 3000 \
  -H "X-Internal-Auth:$(echo -n 'secret' | base64)" \
  -H X-Request-Source:tunnel
```

## Troubleshooting

### "Blocked host" Error (Rails)

**Problem:** Rails shows "Blocked host: tunnel.account.seth4242.net"

**Solution:**
```bash
snet start 3000 --host-header localhost:3000
```

Or add to your Rails config:
```ruby
# config/environments/development.rb
config.hosts << /.*\.seth4242\.net/
```

### CORS Errors

**Problem:** Browser shows CORS errors

**Solution:**
```bash
snet start 8080 \
  --response-header "Access-Control-Allow-Origin:*" \
  --response-header "Access-Control-Allow-Methods:GET,POST,PUT,DELETE,OPTIONS"
```

### Shell Error: "no matches found"

**Problem:** `zsh: no matches found: Access-Control-Allow-Origin:*`

**Solution:** Quote the header value:
```bash
# ❌ Wrong - shell interprets * as glob
--response-header Access-Control-Allow-Origin:*

# ✅ Correct - quoted to prevent expansion
--response-header "Access-Control-Allow-Origin:*"
```

### Custom Domain Not Working

**Problem:** App expects specific host header

**Solution:**
```bash
snet start 3000 --host-header myapp.local
```

### Headers Not Appearing

**Problem:** Custom headers not showing in app

**Verify:**
1. Check flag syntax: `-H name:value` (colon, not equals)
2. **Quote special characters:** `-H "X-Custom:*"` not `-H X-Custom:*`
3. Check for typos in header name
4. Test with curl to verify headers are set
5. Check if your app is reading headers correctly

## Further Reading

- [FRP Documentation](https://github.com/fatedier/frp)
- [HTTP Headers Reference](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers)
- [X-Forwarded-For Standard](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For)
