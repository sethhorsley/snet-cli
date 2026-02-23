# Error Handling Test Guide

This guide shows how to test the improved error handling in the snet CLI.

## Features Added

### 1. Panic Recovery
- **What**: Global panic recovery that catches unexpected crashes
- **Benefit**: Terminal is always restored to normal state (cursor visible, colors reset)
- **Test**: Set `SNET_DEBUG=1` to see detailed stack traces on panic

### 2. Terminal Cleanup
- **What**: Ensures terminal is properly cleaned up on all exit paths
- **Actions Taken**:
  - Show cursor: `\033[?25h`
  - Reset colors: `\033[0m`
  - Exit alt screen: `\033[?1049l`
- **Test**: Kill the CLI with Ctrl+C or let it crash - terminal should be usable

### 3. Formatted Error Messages
- **What**: User-friendly error messages for common failures
- **Errors Handled**:
  - Connection reset/refused → Network troubleshooting steps
  - Authentication failures → Credential troubleshooting steps
  - Generic errors → Clean error display

## Testing

### Test 1: Normal Connection (Should Work Now)
```bash
cd /Users/send16/files/sethhorsley/snet-cli
./bin/snet http 3000 --name test-tunnel
```

**Expected**: CLI connects successfully to FRP server at `snet-frp.fly.dev:7000`

### Test 2: Connection Error with Nice Formatting
```bash
# Temporarily break connectivity (simulate server down)
# You'll see a nicely formatted error instead of a crash
./bin/snet http 3000 --name test-tunnel
```

**Expected Output**:
```
Connection Error

The FRP server is unreachable. This could mean:
  • The server is temporarily down
  • Your network connection is blocking the port
  • The server address has changed

Try these steps:
  1. Check your internet connection
  2. Verify the FRP server is running
  3. Contact support if the issue persists
```

### Test 3: Panic Recovery with Debug Info
```bash
# Set debug mode to see stack traces
SNET_DEBUG=1 ./bin/snet http 3000 --name test-tunnel
```

**Expected**: If a panic occurs, you'll see:
- Clean error message
- Full stack trace (in debug mode)
- Cursor still visible
- Terminal colors reset

### Test 4: TUI Crash Recovery
```bash
# Normal TUI mode - even if TUI crashes, terminal will be restored
./bin/snet http 3000 --name test-tunnel
# Press Ctrl+C or let it crash
```

**Expected**: 
- Terminal remains usable
- No broken terminal state
- Cursor visible
- Can scroll up in history

## Environment Variables

- `SNET_DEBUG=1` - Enable detailed stack traces and debug output

## Verification Checklist

After running the CLI:
- [ ] Can you see the cursor?
- [ ] Are terminal colors normal?
- [ ] Can you scroll up in your shell history?
- [ ] Can you type commands normally?
- [ ] Did error messages look user-friendly?

If all checks pass, the error handling is working correctly!
