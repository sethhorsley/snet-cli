package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "snet",
	Short: "Create secure HTTPS tunnels to localhost",
	Long: `snet creates secure HTTPS tunnels from your local development server
to public URLs on seth4242.net using Cloudflare Tunnels.

Example:
  snet start --port 3000
  
This will create a tunnel and give you a URL like:
  https://abc123.youraccount.seth4242.net`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.snet/config.json)")
}
