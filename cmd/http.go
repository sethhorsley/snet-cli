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
	"github.com/seth4242/snet/internal/errorhandler"
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
	Args:         cobra.MaximumNArgs(1),
	RunE:         runHTTP,
	SilenceUsage: true, // Don't show usage/help on errors
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

		if Verbose {
			fmt.Fprintf(os.Stderr, "\n[DEBUG] Reconnecting to existing tunnel:\n")
			fmt.Fprintf(os.Stderr, "  ID:              %s\n", t.ID)
			fmt.Fprintf(os.Stderr, "  Name:            %s\n", t.Name)
			fmt.Fprintf(os.Stderr, "  URL:             %s\n", t.URL)
			fmt.Fprintf(os.Stderr, "  Provider:        %s\n", t.Provider)
			fmt.Fprintf(os.Stderr, "  FRP Server:      %s:%d\n", t.FRPServerAddr, t.FRPServerPort)
			fmt.Fprintf(os.Stderr, "  FRP Proxy Name:  %s\n", t.FRPProxyName)
			fmt.Fprintf(os.Stderr, "  FRP Auth Token:  %s...\n", maskToken(t.FRPAuthToken))
			fmt.Fprintf(os.Stderr, "\n")
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

		if Verbose {
			fmt.Fprintf(os.Stderr, "\n[DEBUG] Creating tunnel with parameters:\n")
			fmt.Fprintf(os.Stderr, "  Name:       %s\n", tunnelName)
			fmt.Fprintf(os.Stderr, "  Port:       %d\n", port)
			fmt.Fprintf(os.Stderr, "  Wildcard:   %v\n", wildcard)
			fmt.Fprintf(os.Stderr, "  Persistent: %v\n", !httpEphemeral)
			fmt.Fprintf(os.Stderr, "  Provider:   %s\n", provider)
			fmt.Fprintf(os.Stderr, "  API:        %s\n", GetAPIBase())
			fmt.Fprintf(os.Stderr, "\n")
		}

		t, err = client.CreateTunnel(&api.CreateTunnelRequest{
			Name:       tunnelName,
			Port:       port,
			Wildcard:   wildcard,
			Persistent: !httpEphemeral,
			Provider:   provider,
		})
		if err != nil {
			if Verbose {
				fmt.Fprintf(os.Stderr, "\n[DEBUG] Tunnel creation failed:\n")
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "\n")
				fmt.Fprintf(os.Stderr, "[EXPLANATION]\n")
				fmt.Fprintf(os.Stderr, "This error occurs on the Rails API server during tunnel provisioning.\n")
				fmt.Fprintf(os.Stderr, "\n")
				fmt.Fprintf(os.Stderr, "The Rails server follows these steps:\n")
				fmt.Fprintf(os.Stderr, "  1. Generate FRP authentication token (from Rails credentials)\n")
				fmt.Fprintf(os.Stderr, "  2. Create DNS records in Cloudflare (*.%s.%s.snet-public.com)\n", tunnelName, "ACCOUNT")
				fmt.Fprintf(os.Stderr, "  3. Request SSL certificates from Fly.io\n")
				fmt.Fprintf(os.Stderr, "\n")
				fmt.Fprintf(os.Stderr, "The 'Authentication error' indicates step 2 failed.\n")
				fmt.Fprintf(os.Stderr, "This means the Cloudflare API token in Rails credentials is invalid/expired.\n")
				fmt.Fprintf(os.Stderr, "\n")
				fmt.Fprintf(os.Stderr, "To fix this:\n")
				fmt.Fprintf(os.Stderr, "  1. Generate a new Cloudflare API token with Zone.DNS (Edit) permissions\n")
				fmt.Fprintf(os.Stderr, "  2. Update Rails credentials: bin/rails credentials:edit\n")
				fmt.Fprintf(os.Stderr, "  3. Update the cloudflare.api_token value\n")
				fmt.Fprintf(os.Stderr, "\n")
			}
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

		regionName := "iad"              // Default/only region for now
		latency := 50 * time.Millisecond // Placeholder

		// Set region and version for web inspector
		runner.SetRegionAndVersion(regionName, buildinfo.Version)

		// Enable TUI if requested
		if UseTUI && !Quiet {
			if buildinfo.IsDevelopment() {
				totalTime := time.Since(startTime)
				fmt.Fprintf(os.Stderr, "⏱️  Total time to tunnel ready: %v\n", totalTime)
				fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			}

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
			localURL := fmt.Sprintf("http://%s:%d", host, port)
			tuiWrapper := tui.NewTUIWrapper(
				tunnelName,
				t.URL,
				t.WildcardURL,
				localURL,
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

			// Give TUI time to fully clean up before printing error
			time.Sleep(100 * time.Millisecond)

			if err != nil {
				// Check if this is a token mismatch error and we're reconnecting
				if isTokenMismatchError(err) && reconnecting {
					// Force terminal to a completely clean state
					fmt.Print("\033[2K\r")
					fmt.Println()

					// Automatically recreate the tunnel with fresh credentials
					newTunnel, recreateErr := recreateTunnelWithFreshCredentials(
						client, t, port, t.Wildcard, t.Persistent,
					)
					if recreateErr != nil {
						fmt.Printf("Failed to recreate tunnel: %v\n", recreateErr)
						fmt.Println(errorhandler.FormatConnectionError(err))
						return nil
					}

					// Update tunnel reference and restart without TUI (to avoid complexity)
					t = newTunnel
					fmt.Println("Restarting tunnel with fresh credentials...")
					fmt.Println("Press Ctrl+C to disconnect.\n")

					runner = tunnel.NewFRPRunner(client, t, port, host, headerConfig)
					runner.SetRegionAndVersion("iad", buildinfo.Version)

					err = runner.Run()
					if err != nil {
						fmt.Println()
						fmt.Println(errorhandler.FormatConnectionError(err))
					}
					return nil
				}

				// Force terminal to a completely clean state
				// Clear the line and move cursor to start
				fmt.Print("\033[2K\r")

				// Print error to stdout (not stderr) after TUI cleanup
				fmt.Println()
				fmt.Println(errorhandler.FormatConnectionError(err))
				// Return nil to prevent cobra from printing the error again
				return nil
			}
			return nil
		}

		// Plain text mode
		if !Quiet && !reconnecting {
			fmt.Println("Press Ctrl+C to stop the tunnel.\n")
		} else if !Quiet {
			fmt.Println("Press Ctrl+C to disconnect.\n")
		}

		err := runner.Run()
		if err != nil {
			// Check if this is a token mismatch error and we're reconnecting
			if isTokenMismatchError(err) && reconnecting {
				// Automatically recreate the tunnel with fresh credentials
				newTunnel, recreateErr := recreateTunnelWithFreshCredentials(
					client, t, port, t.Wildcard, t.Persistent,
				)
				if recreateErr != nil {
					fmt.Fprintf(os.Stderr, "\nFailed to recreate tunnel: %v\n", recreateErr)
					return err // Return original error
				}

				// Try again with the new tunnel
				t = newTunnel
				fmt.Println("Press Ctrl+C to disconnect.\n")
				runner = tunnel.NewFRPRunner(client, t, port, host, headerConfig)
				runner.SetRegionAndVersion("iad", buildinfo.Version)

				err = runner.Run()
				if err != nil {
					fmt.Fprint(os.Stderr, "\n")
					fmt.Fprintln(os.Stderr, errorhandler.FormatConnectionError(err))
					return err
				}
				return nil
			}

			// Ensure we're on a clean line
			fmt.Fprint(os.Stderr, "\n")
			// Format FRP connection errors nicely
			fmt.Fprintln(os.Stderr, errorhandler.FormatConnectionError(err))
		}
		return err
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

// maskToken masks a token for display, showing only first/last 4 chars
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
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

// isTokenMismatchError checks if an error is caused by FRP token mismatch
func isTokenMismatchError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "token in login doesn't match") ||
		strings.Contains(errMsg, "token mismatch")
}

// recreateTunnelWithFreshCredentials deletes and recreates a tunnel to get fresh credentials
func recreateTunnelWithFreshCredentials(client *api.Client, oldTunnel *api.Tunnel, port int, wildcard bool, persistent bool) (*api.Tunnel, error) {
	fmt.Println("\n⚠️  Detected token mismatch - refreshing tunnel credentials...")

	// Delete the old tunnel
	fmt.Printf("   → Deleting tunnel '%s'...\n", oldTunnel.Name)
	if err := client.DeleteTunnel(oldTunnel.ID); err != nil {
		return nil, fmt.Errorf("failed to delete tunnel: %w", err)
	}

	// Wait a moment for cleanup
	time.Sleep(500 * time.Millisecond)

	// Create a new tunnel with the same name
	fmt.Printf("   → Creating new tunnel with fresh credentials...\n")
	newTunnel, err := client.CreateTunnel(&api.CreateTunnelRequest{
		Name:       oldTunnel.Name,
		Port:       port,
		Wildcard:   wildcard,
		Persistent: persistent,
		Provider:   oldTunnel.Provider,
	})
	if err != nil {
		// Check if this is a server-side authentication issue
		if strings.Contains(err.Error(), "Authentication error") {
			return nil, fmt.Errorf("failed to create tunnel: %w\n\nThis appears to be a server-side issue. The FRP server authentication is not configured correctly.\nPlease contact the system administrator to check the FRP server configuration.", err)
		}
		return nil, fmt.Errorf("failed to create tunnel: %w", err)
	}

	fmt.Println("✓ Tunnel recreated with fresh credentials!\n")
	return newTunnel, nil
}
