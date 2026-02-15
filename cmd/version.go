package cmd

import (
	"fmt"
	"runtime"

	"github.com/seth4242/snet/internal/buildinfo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("snet %s\n", buildinfo.Version)
		fmt.Printf("  mode: %s\n", buildinfo.Mode)
		fmt.Printf("  api: %s\n", buildinfo.GetAPIBase())
		fmt.Printf("  go: %s\n", runtime.Version())
		fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
