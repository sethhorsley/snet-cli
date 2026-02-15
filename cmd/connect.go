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
)

var connectCmd = &cobra.Command{
	Use:   "connect [port]",
	Short: "Connect to an existing persistent tunnel",
	Long: `Connect to an existing persistent tunnel that was created with --persistent.

Example:
  snet connect --tunnel tun_abc123              # connect on port 3000
  snet connect --tunnel tun_abc123 8080         # connect on port 8080
  snet connect -t tun_abc123 --host 192.168.1.5 # tunnel to remote host`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVarP(&connectTunnelID, "tunnel", "t", "", "Tunnel ID to connect to (required)")
	connectCmd.Flags().IntVarP(&connectPort, "port", "p", 0, "Local port to tunnel to (default 3000)")
	connectCmd.Flags().StringVar(&connectHost, "host", "localhost", "Host to tunnel to")
	connectCmd.MarkFlagRequired("tunnel")
}

func runConnect(cmd *cobra.Command, args []string) error {
	// Determine port: positional arg > --port flag > default (3000)
	port := 3000
	if len(args) > 0 {
		p, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[0])
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

	// Verify authentication
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

	// Get tunnel details
	if !Quiet {
		fmt.Println("Fetching tunnel details...")
	}

	t, err := client.GetTunnel(connectTunnelID)
	if err != nil {
		return fmt.Errorf("failed to get tunnel: %w", err)
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

	// Run the appropriate tunnel runner
	if isFRP {
		runner := tunnel.NewFRPRunner(client, t, port, connectHost)
		return runner.Run()
	} else {
		runner := tunnel.NewRunner(client, t, port, connectHost)
		return runner.Run()
	}
}
