package cmd

import (
	"fmt"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/config"
	"github.com/seth4242/snet/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	connectTunnelID string
	connectPort     int
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an existing persistent tunnel",
	Long: `Connect to an existing persistent tunnel that was created with --persistent.

Example:
  snet connect --tunnel tun_abc123 --port 3000`,
	RunE: runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVarP(&connectTunnelID, "tunnel", "t", "", "Tunnel ID to connect to (required)")
	connectCmd.Flags().IntVarP(&connectPort, "port", "p", 3000, "Local port to tunnel to")
	connectCmd.MarkFlagRequired("tunnel")
}

func runConnect(cmd *cobra.Command, args []string) error {
	// Check cloudflared is installed first
	if err := tunnel.CheckCloudflared(); err != nil {
		return err
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Get tunnel details
	fmt.Println("Fetching tunnel details...")

	t, err := client.GetTunnel(connectTunnelID)
	if err != nil {
		return fmt.Errorf("failed to get tunnel: %w", err)
	}

	if t.CloudflareTunnelToken == "" {
		return fmt.Errorf("tunnel does not have a valid token. It may need to be recreated.")
	}

	fmt.Println("Connecting to tunnel...")
	fmt.Println("")
	fmt.Printf("Ready! Proxying localhost:%d to:\n", connectPort)
	fmt.Printf("  %s\n", t.URL)
	if t.WildcardURL != "" {
		fmt.Printf("  %s (wildcard)\n", t.WildcardURL)
	}
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop the tunnel.")
	fmt.Println("")

	// Run the tunnel
	runner := tunnel.NewRunner(client, t, connectPort)
	return runner.Run()
}
