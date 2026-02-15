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
	startPort       int
	startHost       string
	startSubdomain  string
	startWildcard   bool
	startNoWildcard bool
	startPersistent bool
	startName       string
	startProvider   string
)

var startCmd = &cobra.Command{
	Use:   "start [port]",
	Short: "Start a new tunnel",
	Long: `Create a new tunnel and start proxying traffic from seth4242.net to localhost.

Wildcard subdomains are ENABLED BY DEFAULT, allowing *.tunnel.account.seth4242.net

Example:
  snet start                              # tunnel with wildcards (default)
  snet start 8080                         # tunnel to localhost:8080 with wildcards
  snet start 3000 --no-wildcard           # disable wildcard subdomains
  snet start 3000 --subdomain myapp       # request myapp.account.seth4242.net
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
	startCmd.Flags().BoolVarP(&startWildcard, "wildcard", "w", true, "Enable wildcard subdomains (default)")
	startCmd.Flags().BoolVar(&startNoWildcard, "no-wildcard", false, "Disable wildcard subdomains")
	startCmd.Flags().BoolVar(&startPersistent, "persistent", false, "Keep tunnel after disconnect")
	startCmd.Flags().StringVarP(&startName, "name", "n", "", "Friendly name for the tunnel")
	startCmd.Flags().StringVar(&startProvider, "provider", "", "Tunnel provider: frp or cloudflare (uses config default if not specified)")
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

	// Determine wildcard setting
	// Priority: --no-wildcard flag > --wildcard flag > config default > true
	wildcard := cfg.WildcardDefault() // Use config default (defaults to true)
	if cmd.Flags().Changed("wildcard") {
		wildcard = startWildcard
	}
	if startNoWildcard {
		wildcard = false
	}

	// Create tunnel
	if !Quiet {
		fmt.Printf("Creating %s tunnel...\n", provider)
	}

	t, err := client.CreateTunnel(&api.CreateTunnelRequest{
		Name:       startName,
		Subdomain:  startSubdomain,
		Port:       port,
		Wildcard:   wildcard,
		Persistent: startPersistent,
		Provider:   provider,
	})
	if err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}

	if !Quiet {
		fmt.Println("Tunnel created successfully!")
		fmt.Println("")
		fmt.Printf("Ready! Proxying %s:%d to:\n", startHost, port)
		fmt.Printf("  %s\n", t.URL)
		if t.WildcardURL != "" {
			fmt.Printf("  %s (wildcard)\n", t.WildcardURL)
		}
		fmt.Println("")
		fmt.Println("Press Ctrl+C to stop the tunnel.")
		fmt.Println("")
	} else {
		// In quiet mode, just print the URL
		fmt.Println(t.URL)
	}

	// Run the appropriate tunnel runner based on provider
	if t.Provider == "frp" || t.FRPAuthToken != "" {
		runner := tunnel.NewFRPRunner(client, t, port, startHost)
		return runner.Run()
	} else {
		runner := tunnel.NewRunner(client, t, port, startHost)
		return runner.Run()
	}
}
