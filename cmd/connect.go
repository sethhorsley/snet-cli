package cmd

import (
	"fmt"
	"strconv"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/config"
	"github.com/seth4242/snet/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	connectTunnelID string
	connectPort     int
	connectHost     string
	// Header configuration flags
	connectRequestHeaders  []string
	connectResponseHeaders []string
	connectHostHeader      string
)

var connectCmd = &cobra.Command{
	Use:   "connect [name | id] [port]",
	Short: "Attach to an existing tunnel by name or ID",
	Long: `Attach to an existing persistent tunnel by name or tunnel ID.

Tunnels are identified by name (e.g., "api", "acme/api") or by their
tunnel ID (e.g., "tun_abc123").`,
	Example: `  # connect by name
  snet connect api
  snet connect acme/api

  # connect by tunnel ID
  snet connect tun_abc123

  # specify port
  snet connect api 8080`,
	Args: cobra.MinimumNArgs(1),
	RunE: runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVarP(&connectTunnelID, "tunnel", "t", "", "Tunnel ID to connect to (deprecated: use positional argument)")
	connectCmd.Flags().IntVarP(&connectPort, "port", "p", 0, "Local port to tunnel to (default 3000)")
	connectCmd.Flags().StringVar(&connectHost, "host", "localhost", "Host to tunnel to")

	// Header configuration flags
	connectCmd.Flags().StringArrayVarP(&connectRequestHeaders, "header", "H", []string{}, "Add request header (format: name:value, can be repeated)")
	connectCmd.Flags().StringArrayVar(&connectResponseHeaders, "response-header", []string{}, "Add response header (format: name:value, can be repeated)")
	connectCmd.Flags().StringVar(&connectHostHeader, "host-header", "", "Rewrite Host header to specific value")
}

func runConnect(cmd *cobra.Command, args []string) error {
	// First arg is tunnel name/ID, second arg (optional) is port
	tunnelIdentifier := connectTunnelID // From --tunnel flag (deprecated)
	if len(args) > 0 {
		tunnelIdentifier = args[0]
	}

	if tunnelIdentifier == "" {
		return fmt.Errorf("tunnel name or ID required")
	}

	// Determine port: second positional arg > --port flag > default (3000)
	port := 3000
	if len(args) > 1 {
		p, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[1])
		}
		port = p
	} else if connectPort > 0 {
		port = connectPort
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Try to find tunnel by name or ID
	var t *api.Tunnel

	// Check if it's a tunnel ID (starts with "tun_")
	if len(tunnelIdentifier) > 4 && tunnelIdentifier[:4] == "tun_" {
		// It's a tunnel ID
		if !Quiet {
			fmt.Println("Fetching tunnel details...")
		}
		t, err = client.GetTunnel(tunnelIdentifier)
		if err != nil {
			return fmt.Errorf("failed to get tunnel: %w", err)
		}
	} else {
		// It's a tunnel name - search for it
		t, _ = findTunnelByName(client, tunnelIdentifier)
		if t == nil {
			return fmt.Errorf("tunnel '%s' not found", tunnelIdentifier)
		}
	}

	// Determine provider and validate
	isFRP := t.Provider == "frp" || t.FRPAuthToken != ""
	isCloudflare := t.Provider == "cloudflare" || t.CloudflareTunnelToken != ""

	if !isFRP && !isCloudflare {
		return fmt.Errorf("tunnel does not have valid credentials. It may need to be recreated.")
	}

	// Check cloudflared if needed
	if isCloudflare && !isFRP {
		if err := tunnel.CheckCloudflared(); err != nil {
			return err
		}
	}

	if !Quiet {
		fmt.Printf("Connecting to %s tunnel...\n", t.Provider)
		fmt.Println("")
		fmt.Printf("Ready! Proxying %s:%d to:\n", connectHost, port)
		fmt.Printf("  %s\n", t.URL)
		if t.WildcardURL != "" {
			fmt.Printf("  %s (wildcard)\n", t.WildcardURL)
		}
		fmt.Println("")
		fmt.Println("Press Ctrl+C to stop the tunnel.")
		fmt.Println("")
	} else {
		fmt.Println(t.URL)
	}

	// Parse header configuration
	headerConfig, err := tunnel.ParseHeaderFlags(
		connectRequestHeaders,
		connectResponseHeaders,
		connectHostHeader,
	)
	if err != nil {
		return fmt.Errorf("invalid header configuration: %w", err)
	}

	// Run the appropriate tunnel runner
	if isFRP {
		runner := tunnel.NewFRPRunner(client, t, port, connectHost, headerConfig)
		return runner.Run()
	} else {
		runner := tunnel.NewRunner(client, t, port, connectHost)
		return runner.Run()
	}
}
