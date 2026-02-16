# AGENTS.md - Developer Guide for AI Coding Agents

This guide provides essential information for AI coding agents working on the snet-cli codebase.

## Project Overview

- **Language**: Go 1.25.0
- **Module**: `github.com/seth4242/snet`
- **Type**: CLI application for secure tunneling (ngrok-like)
- **Architecture**: Cobra-based CLI with Bubble Tea TUI, supporting Cloudflare and FRP tunnel providers
- **Build System**: Makefile + Go toolchain with dev/prod build modes

## Build Commands

### Development Build (connects to localhost:3001)
```bash
make build          # Build to bin/snet
make install        # Install to $GOPATH/bin
go build -o bin/snet .  # Direct build without ldflags
```

### Production Build (connects to seth4242.net)
```bash
make build-prod     # Build to bin/snet
make install-prod   # Install to $GOPATH/bin
```

### Multi-platform Release
```bash
make release        # Build for macOS, Linux, Windows (amd64/arm64)
```

### Clean
```bash
make clean          # Remove bin/ directory
```

## Test Commands

### Run All Tests
```bash
make test           # Runs: go test -v ./...
go test ./...       # Short output
go test -v ./...    # Verbose output
```

### Run Single Test
```bash
go test -v ./internal/api -run TestFunctionName
go test -v ./cmd -run TestStartCommand
```

### Run Tests with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Note**: Currently NO tests exist in this codebase. When adding tests, create `*_test.go` files alongside source files.

## Linting

**Note**: No linting configuration currently exists. When adding linting, use:

```bash
# Install golangci-lint (recommended)
brew install golangci-lint  # macOS
# or: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
golangci-lint run --fix  # Auto-fix issues
```

## Code Style Guidelines

### Import Organization
Always group imports in three sections with blank lines:
```go
import (
    // 1. Standard library (alphabetical)
    "context"
    "fmt"
    "os"
    
    // 2. External dependencies (alphabetical)
    "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"
    
    // 3. Internal packages (alphabetical)
    "github.com/seth4242/snet/internal/api"
    "github.com/seth4242/snet/internal/config"
)
```

### Naming Conventions
- **Exported** (public): `PascalCase` - `NewClient`, `Client`, `ListTunnels`
- **Unexported** (private): `camelCase` - `runStart`, `apiClient`, `tunnelID`
- **Constants**: `PascalCase` or `SCREAMING_SNAKE_CASE` - `ModeDevelopment`, `DevAPIBase`
- **Package-level vars**: `camelCase` - `startPort`, `startHost`
- **Acronyms**: Keep case consistent - `APIClient`, `HTTPProxy`, `URL` (not `ApiClient`, `HttpProxy`, `Url`)

### Type Definitions
```go
// Structs: Document exported structs, use struct tags
type Tunnel struct {
    ID       string    `json:"id"`
    Name     string    `json:"name"`
    Status   string    `json:"status"`
    Provider string    `json:"provider"` // "cloudflare" or "frp"
}

// Constructors: Always return pointers, use New prefix
func NewClient(cfg *config.Config) *Client {
    return &Client{
        baseURL: cfg.APIBase,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

### Error Handling

**Always wrap errors with context using `fmt.Errorf` and `%w`:**
```go
if err != nil {
    return fmt.Errorf("failed to create tunnel: %w", err)
}
```

**Distinguish fatal vs non-fatal errors:**
```go
// Fatal - return error
if err := config.Load(); err != nil {
    return fmt.Errorf("failed to load config: %w\n\nPlease run 'snet login' first", err)
}

// Non-fatal - log warning and continue
if err != nil {
    fmt.Fprintf(os.Stderr, "Warning: failed to report connection: %v\n", err)
}
```

**Always clean up resources with defer:**
```go
resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("request failed: %w", err)
}
defer resp.Body.Close()  // Always cleanup
```

**Use context for timeouts and cancellation:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### File Organization Pattern
Every Go file should follow this structure:
```go
// 1. Package declaration
package cmd

// 2. Imports (grouped as above)
import (...)

// 3. Package-level variables (for flags, configs, etc.)
var (
    startPort int
    startHost string
)

// 4. Type definitions
type Runner struct {...}

// 5. init() function (for Cobra command registration)
func init() {
    rootCmd.AddCommand(startCmd)
}

// 6. Exported functions
func RunTunnel(...) error {...}

// 7. Unexported helper functions
func runStart(...) error {...}
```

### Cobra Command Pattern
```go
var commandCmd = &cobra.Command{
    Use:   "command [args]",
    Short: "Brief description",
    Long:  `Detailed multi-line description with examples`,
    Args:  cobra.MaximumNArgs(1),
    RunE:  runCommand,  // Use RunE for error returns
}

func init() {
    rootCmd.AddCommand(commandCmd)
    commandCmd.Flags().StringVarP(&flagVar, "flag", "f", "default", "Description")
}

func runCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

### Formatting
- Use `gofmt` (automatic with most editors)
- Indentation: **tabs** (Go standard)
- Line length: No hard limit, but aim for readability (~100-120 chars)
- Comments: Full sentences with proper punctuation
- Add blank lines between logical sections

## Directory Structure

```
snet-cli/
├── main.go                 # Minimal entry point (calls cmd.Execute())
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command + global flags
│   ├── start.go           # Create/start tunnels
│   ├── connect.go         # Connect to existing tunnels
│   ├── list.go            # List tunnels
│   ├── login.go           # Authentication
│   └── ...                # Other commands
├── internal/              # Internal packages (not importable)
│   ├── api/              # API client for seth4242.net
│   ├── buildinfo/        # Build metadata (version, mode)
│   ├── config/           # Config management (~/.snet/config.json)
│   ├── region/           # Region detection (Fly.io)
│   ├── tui/              # Terminal UI (Bubble Tea)
│   └── tunnel/           # Tunnel runners (Cloudflare, FRP)
├── embedded/              # Embedded resources
│   └── binaries/         # FRP binaries (embedded at build time)
├── scripts/              # Build scripts
└── Makefile              # Build automation
```

**Key Principles:**
- `cmd/` contains only CLI command definitions
- `internal/` contains all business logic
- Use `internal/` to prevent external imports
- Each package has a single, clear responsibility

## Adding New Commands

1. Create `cmd/newcommand.go`
2. Define command with `cobra.Command`
3. Register in `init()` with `rootCmd.AddCommand(newCmd)`
4. Implement `RunE` function for logic
5. Add business logic to appropriate `internal/` package

## Configuration

- Config file: `~/.snet/config.json`
- Loaded via `config.Load()`
- Contains: API token, account ID, API base URL, default provider
- Never commit real tokens (see `.gitignore`)

## Build Modes

Two modes controlled by ldflags in Makefile:
- **Development**: `Mode=development`, uses `http://localhost:3001/api/v1`
- **Production**: `Mode=production`, uses `https://seth4242.net/api/v1`

Access via `buildinfo.Mode` and `buildinfo.Version`

## Dependencies

Key libraries:
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - Terminal UI
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/fatedier/frp` - FRP tunnel client (embedded)

## Common Patterns

### Graceful Shutdown
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

select {
case <-sigChan:
    fmt.Println("\nShutting down...")
    cleanup()
case err := <-done:
    // Handle completion
}
```

### API Error Handling
```go
type APIError struct {
    Error  string   `json:"error,omitempty"`
    Errors []string `json:"errors,omitempty"`
}

func (e *APIError) String() string {
    if e.Error != "" {
        return e.Error
    }
    if len(e.Errors) > 0 {
        return e.Errors[0]
    }
    return "unknown error"
}
```

## Version Control

- Use semantic versioning: `v1.2.3`
- Version derived from git tags: `git describe --tags --always --dirty`
- Never commit binaries (see `.gitignore`)

## Important Notes for Agents

1. **Never commit without explicit instruction** - Follow guidance from ~/.config/Claude/AGENTS.md
2. **Always wrap errors** - Use `fmt.Errorf("context: %w", err)` pattern
3. **Use internal/ packages** - Keep business logic out of cmd/
4. **Follow import grouping** - Standard lib, external, internal
5. **Clean up resources** - Use defer for Close(), cleanup, etc.
6. **Test manually** - No automated tests exist yet
7. **Check both build modes** - Ensure code works in dev and prod
8. **Document exported types** - Add comments for public APIs
