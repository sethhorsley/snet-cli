package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/config"
	"github.com/seth4242/snet/internal/region"
	"github.com/seth4242/snet/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	startPort       int
	startHost       string
	startSubdomain  string
	startPersistent bool
	startWildcard   bool
	startName       string
	startProvider   string
	// Header configuration flags
	startRequestHeaders  []string
	startResponseHeaders []string
	startHostHeader      string
)

var startCmd = &cobra.Command{
	Use:   "start [port]",
	Short: "Start a new tunnel",
	Long: `Create a new tunnel and start proxying traffic from snet-public.com to localhost.

Tunnels are covered by wildcard SSL certificate (*.snet-public.com).
URL pattern: tunnel-slug--account-slug.snet-public.com

With --wildcard flag (requires account wildcard enabled):
  - All subdomains route to your tunnel: *.account.snet-public.com
  - Example: api.account.snet-public.com, admin.account.snet-public.com

Example:
  snet start                              # tunnel to localhost:3000
  snet start 8080                         # tunnel to localhost:8080
  snet start 3000 --wildcard              # enable wildcard subdomain routing
  snet start 3000 --subdomain myapp       # request myapp--account.snet-public.com
  snet start 3000 --host 192.168.1.100    # tunnel to remote host
  snet start --persistent --name my-proj  # persistent tunnel with name`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVarP(&startPort, "port", "p", 0, "Local port to tunnel to (default 3000)")
	startCmd.Flags().StringVar(&startHost, "host", "localhost", "Host to tunnel to")
	startCmd.Flags().StringVarP(&startSubdomain, "subdomain", "s", "", "Request a specific subdomain")
	startCmd.Flags().BoolVar(&startPersistent, "persistent", false, "Keep tunnel after disconnect")
	startCmd.Flags().BoolVarP(&startWildcard, "wildcard", "w", false, "Enable wildcard subdomain routing (*.account.snet-public.com)")
	startCmd.Flags().StringVarP(&startName, "name", "n", "", "Friendly name for the tunnel")
	startCmd.Flags().StringVar(&startProvider, "provider", "", "Tunnel provider: frp or cloudflare (uses config default if not specified)")

	// Header configuration flags
	startCmd.Flags().StringArrayVarP(&startRequestHeaders, "header", "H", []string{}, "Add request header (format: name:value, can be repeated)")
	startCmd.Flags().StringArrayVar(&startResponseHeaders, "response-header", []string{}, "Add response header (format: name:value, can be repeated)")
	startCmd.Flags().StringVar(&startHostHeader, "host-header", "", "Rewrite Host header to specific value")
}

func runStart(cmd *cobra.Command, args []string) error {
	// Determine port: positional arg > --port flag > default (3000)
	port := 3000
	if len(args) > 0 {
		p, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[0])
		}
		port = p
	} else if startPort > 0 {
		port = startPort
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Verify authentication by listing accounts
	if !Quiet {
		fmt.Println("Verifying authentication...")
	}
	client := api.NewClient(cfg)
	accounts, err := client.ListAccounts()
	if err != nil {
		return fmt.Errorf("authentication failed: %w\n\nPlease run 'snet login' to authenticate", err)
	}
	if len(accounts) == 0 {
		return fmt.Errorf("no accounts found for this token\n\nPlease run 'snet login' to authenticate")
	}

	// Determine which provider to use
	provider := startProvider
	if provider == "" {
		// User didn't specify --provider, use config default
		provider = cfg.DefaultProvider
		if provider == "" {
			provider = "frp" // Ultimate fallback
		}
	}

	// Validate provider
	if provider != "frp" && provider != "cloudflare" {
		return fmt.Errorf("invalid provider: %s (must be 'frp' or 'cloudflare')", provider)
	}

	// Check cloudflared is installed if using Cloudflare
	if provider == "cloudflare" {
		if err := tunnel.CheckCloudflared(); err != nil {
			return err
		}
	}

	// Detect closest region for FRP tunnels
	var closestRegion string
	if provider == "frp" {
		if !Quiet {
			fmt.Println("Detecting closest regions...")
		}

		// Get available regions from API
		availableResp, err := client.GetRegions()
		if err != nil {
			if !Quiet {
				fmt.Printf("  ⚠ Could not fetch available regions: %v\n", err)
			}
		}

		availableRegions := make(map[string]bool)
		if availableResp != nil {
			for _, r := range availableResp.Regions {
				availableRegions[r.Code] = true
			}
		}

		// Detect closest region using Fly.io edge
		closestRegion, err = region.DetectClosestRegion()
		if err != nil {
			if !Quiet {
				fmt.Printf("  ⚠ Could not detect region, using default: %v\n", err)
			}
			closestRegion = "ord" // Default to Chicago
		}

		// Check if closest region is available
		if availableRegions[closestRegion] {
			if !Quiet {
				fmt.Printf("  ✓ Connecting to %s (%s)\n", closestRegion, region.GetRegionName(closestRegion))
			}
		} else {
			// Closest region not available - measure latency to find alternatives
			if !Quiet {
				fmt.Printf("  → Closest: %s (%s) - not available yet\n", closestRegion, region.GetRegionName(closestRegion))
				fmt.Println("  → Measuring latency to find alternatives...")
			}

			closestRegions, err := region.DetectClosestRegions(5)
			if err == nil && len(closestRegions) > 0 {
				// Submit region request for the closest unavailable region
				if !availableRegions[closestRegions[0].Code] {
					client.RequestRegion(&api.RegionRequestRequest{
						RegionCode:        closestRegions[0].Code,
						DetectedLatencyMS: int(closestRegions[0].LatencyMS),
					})
				}

				if !Quiet {
					fmt.Println("\n  Closest regions by latency:")
					fmt.Println("  ┌────────┬─────────────────────────────┬───────────┬──────────┐")
					fmt.Println("  │ Region │ Location                    │ Latency   │ Status   │")
					fmt.Println("  ├────────┼─────────────────────────────┼───────────┼──────────┤")
					for i, r := range closestRegions {
						if i >= 5 {
							break
						}
						status := "Available"
						if !availableRegions[r.Code] {
							status = "Requested"
						}
						fmt.Printf("  │ %-6s │ %-27s │ %6dms │ %-8s │\n",
							r.Code, r.Name, r.LatencyMS, status)
					}
					fmt.Println("  └────────┴─────────────────────────────┴───────────┴──────────┘")
					fmt.Println()
				}
			}

			// Use first available region or default
			for code := range availableRegions {
				closestRegion = code
				break
			}
			if closestRegion == "" {
				closestRegion = "ord"
			}

			if !Quiet {
				fmt.Printf("  → Using %s (%s) via anycast routing\n", closestRegion, region.GetRegionName(closestRegion))
			}
		}
	}

	// Determine tunnel name - use current directory name if not specified
	tunnelName := startName
	if tunnelName == "" {
		cwd, err := os.Getwd()
		if err == nil {
			tunnelName = filepath.Base(cwd)
		}
		// If we still don't have a name, use a generic one
		if tunnelName == "" || tunnelName == "." || tunnelName == "/" {
			tunnelName = "dev"
		}
	}

	if Verbose {
		fmt.Printf("[DEBUG] Tunnel name: %q\n", tunnelName)
	}

	// Create tunnel
	if !Quiet {
		fmt.Printf("Creating %s tunnel...\n", provider)
		fmt.Println("  → Generating authentication credentials")
		if provider == "frp" {
			if startWildcard {
				fmt.Println("  → Requesting wildcard subdomain routing")
			}
			fmt.Println("  → Using wildcard SSL certificate")
		}
	}

	t, err := client.CreateTunnel(&api.CreateTunnelRequest{
		Name:       tunnelName,
		Subdomain:  startSubdomain,
		Port:       port,
		Persistent: startPersistent,
		Wildcard:   startWildcard,
		Provider:   provider,
		Region:     closestRegion,
	})
	if err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}

	if !Quiet {
		fmt.Println("✓ Tunnel created successfully!")
		fmt.Println("")
		fmt.Printf("Ready! Proxying %s:%d to:\n", startHost, port)
		fmt.Printf("  %s\n", t.URL)
		if t.WildcardEnabled && t.WildcardURL != "" {
			fmt.Printf("  %s (wildcard)\n", t.WildcardURL)
		}
		fmt.Println("")
		fmt.Println("Press Ctrl+C to stop the tunnel.")
		fmt.Println("")
	} else {
		// In quiet mode, just print the URL
		fmt.Println(t.URL)
	}

	// Parse header configuration
	headerConfig, err := tunnel.ParseHeaderFlags(
		startRequestHeaders,
		startResponseHeaders,
		startHostHeader,
	)
	if err != nil {
		return fmt.Errorf("invalid header configuration: %w", err)
	}

	// Run the appropriate tunnel runner based on provider
	if t.Provider == "frp" || t.FRPAuthToken != "" {
		runner := tunnel.NewFRPRunner(client, t, port, startHost, headerConfig)
		return runner.Run()
	} else {
		runner := tunnel.NewRunner(client, t, port, startHost)
		return runner.Run()
	}
}
