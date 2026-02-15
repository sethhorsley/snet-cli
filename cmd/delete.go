package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/config"
	"github.com/spf13/cobra"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:     "delete [tunnel-id]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a tunnel",
	Long: `Delete a tunnel and its associated Cloudflare resources.

Example:
  snet delete tun_abc123
  snet delete tun_abc123 -y
  snet delete tun_abc123 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation")
	deleteCmd.Flags().BoolVarP(&deleteForce, "yes", "y", false, "Skip confirmation (alias for --force)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	tunnelID := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Get tunnel details for confirmation
	t, err := client.GetTunnel(tunnelID)
	if err != nil {
		return fmt.Errorf("failed to get tunnel: %w", err)
	}

	// Confirm deletion
	if !deleteForce {
		fmt.Printf("Are you sure you want to delete tunnel '%s'?\n", t.URL)
		fmt.Printf("This will remove the tunnel and all Cloudflare resources.\n")
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(strings.ToLower(confirmation))

		if confirmation != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete tunnel
	fmt.Println("Deleting tunnel...")

	if err := client.DeleteTunnel(tunnelID); err != nil {
		return fmt.Errorf("failed to delete tunnel: %w", err)
	}

	fmt.Println("Tunnel deleted successfully.")
	return nil
}
