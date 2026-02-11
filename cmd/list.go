package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all tunnels",
	Long:    `List all tunnels for the current account.`,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// List tunnels
	tunnels, err := client.ListTunnels()
	if err != nil {
		return fmt.Errorf("failed to list tunnels: %w", err)
	}

	if len(tunnels) == 0 {
		fmt.Println("No tunnels found.")
		fmt.Println("")
		fmt.Println("Create one with: snet start --port 3000")
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tURL\tSTATUS\tTYPE")
	fmt.Fprintln(w, "--\t----\t---\t------\t----")

	for _, t := range tunnels {
		name := t.Name
		if name == "" {
			name = t.Slug
		}

		tunnelType := "ephemeral"
		if t.Persistent {
			tunnelType = "persistent"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			name,
			t.URL,
			t.Status,
			tunnelType,
		)
	}

	w.Flush()
	return nil
}
