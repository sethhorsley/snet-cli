package cmd

import (
	"fmt"

	"github.com/seth4242/snet/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long:  `Remove stored API token and account configuration.`,
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	if !config.Exists() {
		fmt.Println("Not logged in.")
		return nil
	}

	if err := config.Delete(); err != nil {
		return fmt.Errorf("failed to remove config: %w", err)
	}

	fmt.Println("Logged out successfully.")
	return nil
}
