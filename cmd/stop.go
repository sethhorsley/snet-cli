package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [tunnel-slug-or-pid]",
	Short: "Stop a running tunnel",
	Long: `Stop a running tunnel by slug or process ID.
	
If no argument is provided, attempts to stop all snet tunnel processes.

Example:
  snet stop                    # stop all snet tunnels
  snet stop my-tunnel-abc123   # stop by tunnel slug
  snet stop 12345              # stop by process ID`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// Stop all snet tunnel processes
		return stopAllTunnels()
	}

	arg := args[0]

	// Try as PID first
	if pid, err := strconv.Atoi(arg); err == nil {
		return stopByPID(pid)
	}

	// Try as tunnel slug
	return stopBySlug(arg)
}

// stopAllTunnels stops all running snet tunnel processes
func stopAllTunnels() error {
	// Find all snet processes
	cmd := exec.Command("pgrep", "-f", "snet (start|connect)")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			fmt.Println("No running tunnels found")
			return nil
		}
		return fmt.Errorf("failed to find tunnel processes: %w", err)
	}

	pids := strings.Fields(string(output))
	if len(pids) == 0 {
		fmt.Println("No running tunnels found")
		return nil
	}

	stopped := 0
	for _, pidStr := range pids {
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Don't kill ourselves
		if pid == os.Getpid() {
			continue
		}

		if err := stopByPID(pid); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to stop PID %d: %v\n", pid, err)
		} else {
			stopped++
		}
	}

	if stopped > 0 {
		fmt.Printf("Stopped %d tunnel(s)\n", stopped)
	} else {
		fmt.Println("No tunnels were stopped")
	}

	return nil
}

// stopByPID stops a tunnel by process ID
func stopByPID(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %d", pid)
	}

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Try SIGKILL if SIGTERM fails
		if err := process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to stop process %d: %w", pid, err)
		}
	}

	if !Quiet {
		fmt.Printf("Stopped tunnel (PID: %d)\n", pid)
	}

	return nil
}

// stopBySlug stops a tunnel by finding its PID file
func stopBySlug(slug string) error {
	// Look for PID files in common locations
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pidLocations := []string{
		filepath.Join(home, ".snet", "tunnels", slug+".pid"),
		filepath.Join(home, ".snet", slug+".pid"),
		filepath.Join("/tmp", "snet-"+slug+".pid"),
	}

	for _, pidFile := range pidLocations {
		if data, err := os.ReadFile(pidFile); err == nil {
			pidStr := strings.TrimSpace(string(data))
			if pid, err := strconv.Atoi(pidStr); err == nil {
				if err := stopByPID(pid); err == nil {
					// Clean up PID file
					os.Remove(pidFile)
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no running tunnel found with slug: %s", slug)
}
