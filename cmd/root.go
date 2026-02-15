package cmd

import (
	"fmt"
	"os"

	"github.com/seth4242/snet/internal/buildinfo"
	"github.com/spf13/cobra"
)

var (
	// apiPort overrides the API port (development mode only)
	apiPort int

	// Quiet suppresses non-essential output
	Quiet bool

	// Verbose enables detailed output
	Verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "snet",
	Short: "Create secure HTTPS tunnels to localhost",
	Long: `snet creates secure HTTPS tunnels from your local development server
to public URLs on seth4242.net using Cloudflare Tunnels.

Example:
  snet start 3000
  
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
	rootCmd.PersistentFlags().IntVar(&apiPort, "api-port", 0, "Override API port (development mode only)")
	rootCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
}

// GetAPIBase returns the API base URL, respecting the --api-port flag
func GetAPIBase() string {
	if apiPort > 0 && buildinfo.IsDevelopment() {
		return fmt.Sprintf("http://localhost:%d/api/v1", apiPort)
	}
	return buildinfo.GetAPIBase()
}
