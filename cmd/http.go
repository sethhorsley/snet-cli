package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/buildinfo"
	"github.com/seth4242/snet/internal/config"
	"github.com/seth4242/snet/internal/tui"
	"github.com/seth4242/snet/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	httpName      string
	httpURL       string
	httpInspect   bool
	httpEphemeral bool
	// Header configuration flags
	httpRequestHeaders  []string
	httpResponseHeaders []string
	httpHostHeader      string
)

var httpCmd = &cobra.Command{
	Use:   "http [address:port | port | url]",
	Short: "Start an HTTP tunnel",
	Long: `Forward HTTP traffic from a public HTTPS URL to a local port, address, or URL.

Tunnels are persistent by default. Running this command again reconnects to
the same public URL unless a different name is specified.

Use --ephemeral for temporary tunnels that are deleted on disconnect.`,
	Example: `  # forward traffic to localhost:8080
  snet http 8080

  # explicit name
  snet http 3000 --name api

  # name from current directory
  snet http 3000 --name .

  # forward to another host
  snet http server.local:9000

  # ephemeral tunnel (auto-cleanup on disconnect)
  snet http 3000 --ephemeral
  snet http 3000 --temp         # same as --ephemeral
  snet http 3000 --tmp          # same as --ephemeral

  # choose endpoint URL
  snet http 3000 --url https://api.example.com

  # custom headers for multi-tenant apps
  snet http 3000 --host-header tenant1.example.com -H "X-Forwarded-Host:tenant1.example.com"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHTTP,
}

func init() {
	rootCmd.AddCommand(httpCmd)

	httpCmd.Flags().StringVar(&httpName, "name", "", "tunnel name (default: current directory)")
	httpCmd.Flags().StringVar(&httpURL, "url", "", "request a specific hostname")
	httpCmd.Flags().BoolVar(&httpInspect, "inspect", true, "enable/disable http inspection")
	httpCmd.Flags().BoolVar(&httpEphemeral, "ephemeral", false, "temporary tunnel (auto-cleanup on disconnect)")
	httpCmd.Flags().BoolVar(&httpEphemeral, "temp", false, "temporary tunnel (alias for --ephemeral)")
	httpCmd.Flags().BoolVar(&httpEphemeral, "tmp", false, "temporary tunnel (alias for --ephemeral)")

	// Header configuration flags
	httpCmd.Flags().StringArrayVarP(&httpRequestHeaders, "header", "H", []string{}, "Add request header (format: name:value, can be repeated)")
	httpCmd.Flags().StringArrayVar(&httpResponseHeaders, "response-header", []string{}, "Add response header (format: name:value, can be repeated)")
	httpCmd.Flags().StringVar(&httpHostHeader, "host-header", "", "Rewrite Host header to specific value")
}

func runHTTP(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Parse address:port or just port
	port := 3000
	host := "localhost"

	if len(args) > 0 {
		// Try to parse as port first
		if p, err := strconv.Atoi(args[0]); err == nil {
			port = p
		} else if strings.Contains(args[0], ":") {
			// address:port format
			parts := strings.Split(args[0], ":")
			host = parts[0]
			if p, err := strconv.Atoi(parts[1]); err == nil {
				port = p
			} else {
				return fmt.Errorf("invalid port in %s", args[0])
			}
		} else {
			return fmt.Errorf("invalid argument: %s (expected port or address:port)", args[0])
		}
	}

	// Determine tunnel name
	tunnelName := httpName
	if tunnelName == "" {
		// Default to current directory name
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		tunnelName = filepath.Base(cwd)
	} else if tunnelName == "." {
		// Explicit current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		tunnelName = filepath.Base(cwd)
	}

	// Sanitize tunnel name (lowercase, alphanumeric, hyphens only)
	tunnelName = sanitizeTunnelName(tunnelName)

	if buildinfo.IsDevelopment() {
		fmt.Fprintf(os.Stderr, "⏱️  CLI startup time: %v\n", time.Since(startTime))
	}

	// Load config
	configStart := time.Now()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if buildinfo.IsDevelopment() {
		fmt.Fprintf(os.Stderr, "⏱️  Config load time: %v\n", time.Since(configStart))
	}

	client := api.NewClient(cfg)

	// For TUI mode, we'll start immediately and fetch tunnel data asynchronously
	// For non-TUI mode, we block and wait
	var t *api.Tunnel
	var reconnecting bool

	// Fetch tunnel data
	// For TUI mode, we could show a loading spinner here
	// For now, both modes block on this call
	var existingTunnel *api.Tunnel

	if UseTUI && !Quiet {
		fmt.Fprintf(os.Stderr, "⏳ Loading tunnel data...\n")
	}

	apiStart := time.Now()
	if !httpEphemeral {
		existingTunnel, reconnecting = findTunnelByName(client, tunnelName)
	}
	if buildinfo.IsDevelopment() && existingTunnel != nil {
		fmt.Fprintf(os.Stderr, "⏱️  API call (list tunnels): %v\n", time.Since(apiStart))
	}

	if existingTunnel != nil {
		// Reconnect to existing tunnel
		t = existingTunnel
		if !Quiet {
			fmt.Printf("%s\n", t.URL)
			if t.WildcardURL != "" {
				fmt.Printf("%s (wildcard)\n", t.WildcardURL)
			}
			fmt.Printf("name: %s (reconnected)\n\n", tunnelName)
		}
	} else {
		// Create new tunnel
		provider := cfg.DefaultProvider
		if provider == "" {
			provider = "frp"
		}

		wildcard := cfg.WildcardDefault()

		if !Quiet {
			fmt.Printf("Creating %s tunnel...\n", provider)
			if wildcard {
				fmt.Println("  → Provisioning wildcard support")
			}
			fmt.Println("  → Generating authentication credentials")
		}

		createStart := time.Now()
		t, err = client.CreateTunnel(&api.CreateTunnelRequest{
			Name:       tunnelName,
			Port:       port,
			Wildcard:   wildcard,
			Persistent: !httpEphemeral,
			Provider:   provider,
		})
		if err != nil {
			return fmt.Errorf("failed to create tunnel: %w", err)
		}
		if buildinfo.IsDevelopment() {
			fmt.Fprintf(os.Stderr, "⏱️  API call (create tunnel): %v\n", time.Since(createStart))
		}

		if !Quiet {
			fmt.Println("✓ Tunnel created successfully!\n")
			fmt.Printf("%s\n", t.URL)
			if t.WildcardURL != "" {
				fmt.Printf("%s (wildcard)\n", t.WildcardURL)
			}
			fmt.Printf("name: %s\n\n", tunnelName)
		}
	}

	// Parse header configuration
	headerConfig, err := tunnel.ParseHeaderFlags(
		httpRequestHeaders,
		httpResponseHeaders,
		httpHostHeader,
	)
	if err != nil {
		return fmt.Errorf("invalid header configuration: %w", err)
	}

	// Start the tunnel runner
	if t.Provider == "frp" || t.FRPAuthToken != "" {
		runner := tunnel.NewFRPRunner(client, t, port, host, headerConfig)

		// Enable TUI if requested
		if UseTUI && !Quiet {
			if buildinfo.IsDevelopment() {
				totalTime := time.Since(startTime)
				fmt.Fprintf(os.Stderr, "⏱️  Total time to tunnel ready: %v\n", totalTime)
				fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			}

			regionName := "iad"              // Default/only region for now
			latency := 50 * time.Millisecond // Placeholder

			// Get account name from tunnel
			accountName := t.Account.Name
			if accountName == "" {
				accountName = "default"
			}

			// Get header configuration details
			headerSummary := ""
			hostHeaderRewrite := ""
			var requestHeaders map[string]string
			var responseHeaders map[string]string

			if headerConfig != nil {
				headerSummary = headerConfig.String()
				hostHeaderRewrite = headerConfig.HostHeaderRewrite
				requestHeaders = headerConfig.RequestHeaders
				responseHeaders = headerConfig.ResponseHeaders
			}

			// Create TUI wrapper with real data
			tuiWrapper := tui.NewTUIWrapper(
				tunnelName,
				t.URL,
				t.WildcardURL,
				accountName,
				regionName,
				buildinfo.Version,
				t.Wildcard,
				reconnecting,
				latency,
				headerSummary,
				hostHeaderRewrite,
				requestHeaders,
				responseHeaders,
			)

			// Set TUI callbacks on runner
			runner.SetTUICallbacks(tuiWrapper)

			// Start TUI
			if err := tuiWrapper.Start(); err != nil {
				// Fall back to plain mode
				fmt.Fprintf(os.Stderr, "Warning: TUI failed to start: %v\n", err)
				fmt.Println("\nPress Ctrl+C to stop the tunnel.\n")
				return runner.Run()
			}

			// Run tunnel (blocks until quit)
			err := runner.Run()
			tuiWrapper.Stop()
			return err
		}

		// Plain text mode
		if !Quiet && !reconnecting {
			fmt.Println("Press Ctrl+C to stop the tunnel.\n")
		} else if !Quiet {
			fmt.Println("Press Ctrl+C to disconnect.\n")
		}

		return runner.Run()
	} else {
		runner := tunnel.NewRunner(client, t, port, host)

		if !Quiet && !reconnecting {
			fmt.Println("Press Ctrl+C to stop the tunnel.\n")
		} else if !Quiet {
			fmt.Println("Press Ctrl+C to disconnect.\n")
		}

		return runner.Run()
	}
}

// findTunnelByName searches for an existing tunnel by name
func findTunnelByName(client *api.Client, name string) (*api.Tunnel, bool) {
	tunnels, err := client.ListTunnels()
	if err != nil {
		return nil, false
	}

	for _, t := range tunnels {
		if t.Name == name {
			// Get full tunnel details with token
			fullTunnel, err := client.GetTunnel(t.ID)
			if err != nil {
				return nil, false
			}
			return fullTunnel, true
		}
	}

	return nil, false
}

// sanitizeTunnelName converts a name to lowercase alphanumeric with hyphens
func sanitizeTunnelName(name string) string {
	// Replace underscores with hyphens
	name = strings.ReplaceAll(name, "_", "-")
	// Remove any non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '/' {
			result.WriteRune(r)
		}
	}
	return strings.ToLower(result.String())
}
