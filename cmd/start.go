package cmd

import (
	"fmt"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/config"
	"github.com/seth4242/snet/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	startPort       int
	startWildcard   bool
	startPersistent bool
	startName       string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new tunnel",
	Long: `Create a new tunnel and start proxying traffic from seth4242.net to localhost.

Example:
  snet start --port 3000
  snet start --port 3000 --wildcard
  snet start --port 3000 --persistent --name my-project`,
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVarP(&startPort, "port", "p", 3000, "Local port to tunnel to")
	startCmd.Flags().BoolVarP(&startWildcard, "wildcard", "w", false, "Enable wildcard subdomains")
	startCmd.Flags().BoolVar(&startPersistent, "persistent", false, "Keep tunnel after disconnect")
	startCmd.Flags().StringVarP(&startName, "name", "n", "", "Friendly name for the tunnel")
}

func runStart(cmd *cobra.Command, args []string) error {
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

	// Create tunnel
	fmt.Println("Creating tunnel...")

	t, err := client.CreateTunnel(&api.CreateTunnelRequest{
		Name:       startName,
		Wildcard:   startWildcard,
		Persistent: startPersistent,
	})
	if err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}

	fmt.Println("Tunnel created successfully!")
	fmt.Println("")
	fmt.Printf("Ready! Proxying localhost:%d to:\n", startPort)
	fmt.Printf("  %s\n", t.URL)
	if t.WildcardURL != "" {
		fmt.Printf("  %s (wildcard)\n", t.WildcardURL)
	}
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop the tunnel.")
	fmt.Println("")

	// Run the tunnel
	runner := tunnel.NewRunner(client, t, startPort)
	return runner.Run()
}
