package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/buildinfo"
	"github.com/seth4242/snet/internal/config"
	"github.com/spf13/cobra"
)

var listJSON bool

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all tunnels",
	Long: `List all tunnels for the current account.

Example:
  snet list                # table format
  snet list --json         # JSON for scripting
  snet list --json | jq '.[] | select(.status=="active")'`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
}

func runList(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if buildinfo.IsDevelopment() {
		fmt.Fprintf(os.Stderr, "⏱️  Config load time: %v\n", time.Since(startTime))
	}

	client := api.NewClient(cfg)

	// List tunnels
	apiStart := time.Now()
	tunnels, err := client.ListTunnels()
	if err != nil {
		return fmt.Errorf("failed to list tunnels: %w", err)
	}
	if buildinfo.IsDevelopment() {
		fmt.Fprintf(os.Stderr, "⏱️  API call (list tunnels): %v\n", time.Since(apiStart))
		fmt.Fprintf(os.Stderr, "⏱️  Total time: %v\n", time.Since(startTime))
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	}

	// JSON output
	if listJSON {
		data, err := json.MarshalIndent(tunnels, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Table output
	if len(tunnels) == 0 {
		fmt.Println("No tunnels found.")
		fmt.Println("")
		fmt.Println("Create one with: snet start 3000")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tURL\tSTATUS\tTYPE\tSSL")
	fmt.Fprintln(w, "--\t----\t---\t------\t----\t---")

	for _, t := range tunnels {
		name := t.Name
		if name == "" {
			name = t.Slug
		}

		tunnelType := "ephemeral"
		if t.Persistent {
			tunnelType = "persistent"
		}

		// Determine SSL status display
		sslStatus := "-"
		if t.Wildcard {
			if t.SSLReady {
				sslStatus = "✓"
			} else {
				sslStatus = "⏳"
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			name,
			t.URL,
			t.Status,
			tunnelType,
			sslStatus,
		)
	}

	w.Flush()
	return nil
}
