package cmd

import (
	"fmt"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/config"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show information about the authenticated user",
	Long: `Display information about the current user and their accounts.

Shows:
  - User name and email
  - Current active account
  - All accounts the user belongs to
  - Account ownership status`,
	RunE: runWhoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("not logged in. Run 'snet login' first")
	}

	// Create API client
	client := api.NewClient(cfg)

	// Get user info
	me, err := client.GetMe()
	if err != nil {
		return fmt.Errorf("failed to fetch user info: %w", err)
	}

	// Display user info
	fmt.Println("User Information")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Name:      %s\n", me.User.Name)
	fmt.Printf("Email:     %s\n", me.User.Email)
	if me.User.Admin {
		fmt.Println("Role:      Administrator")
	}
	fmt.Printf("User ID:   %d\n", me.User.ID)

	// Show API token (masked)
	token := cfg.APIToken
	if len(token) > 8 {
		maskedToken := token[:4] + "..." + token[len(token)-4:]
		fmt.Printf("API Token: %s\n", maskedToken)
	}
	fmt.Println()

	// Display current account
	if me.CurrentAccount != nil {
		fmt.Println("Current Account")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Name:      %s\n", me.CurrentAccount.Name)
		fmt.Printf("Slug:      %s\n", me.CurrentAccount.Slug)
		fmt.Printf("ID:        %s\n", me.CurrentAccount.PrefixID)
		if me.CurrentAccount.Personal {
			fmt.Println("Type:      Personal")
		} else {
			fmt.Println("Type:      Team")
		}
		fmt.Println()
	}

	// Display all accounts
	if len(me.Accounts) > 0 {
		fmt.Printf("All Accounts (%d)\n", len(me.Accounts))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for _, account := range me.Accounts {
			ownerStr := ""
			if account.Owner {
				ownerStr = " (owner)"
			}

			typeStr := "team"
			if account.Personal {
				typeStr = "personal"
			}

			fmt.Printf("  • %s (%s)%s\n", account.Name, typeStr, ownerStr)
			fmt.Printf("    Slug: %s\n", account.Slug)
			fmt.Printf("    ID:   %s\n", account.PrefixID)
			fmt.Println()
		}
	}

	return nil
}
