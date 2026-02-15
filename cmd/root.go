package cmd

import (
	"fmt"
	"os"
	"strconv"

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
	Short: "Secure tunnels from local ports to public endpoints",
	Long: `snet - secure tunnels from local ports to public endpoints

USAGE:
  snet [command] [flags]
  snet <port>                  # shortcut for: snet http <port>

QUICK START:
  snet 3000                    # start/reconnect HTTP tunnel named after current directory
  snet http 8080               # explicit HTTP tunnel
  snet connect api             # attach to an existing tunnel by name
  snet list                    # list tunnels and status

DESCRIPTION:
  snet creates persistent tunnels. Re-running the same command reconnects
  to the same public URL unless the name is changed.`,
}

// Execute runs the root command
func Execute() {
	// Check if first arg is a port number before cobra processes it
	if len(os.Args) > 1 {
		if _, err := strconv.Atoi(os.Args[1]); err == nil {
			// It's a port number, insert "http" command
			args := make([]string, len(os.Args)+1)
			args[0] = os.Args[0]
			args[1] = "http"
			copy(args[2:], os.Args[1:])
			os.Args = args
		}
	}

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
