package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/seth4242/snet/internal/api"
	"github.com/seth4242/snet/internal/buildinfo"
	"github.com/seth4242/snet/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with seth4242.net",
	Long: `Login to seth4242.net by providing your API token.

Create an API token at: https://seth4242.net/api_tokens

The token will be stored in ~/.snet/config.json`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	apiBase := GetAPIBase()

	// Show appropriate URL based on mode
	tokenURL := "https://seth4242.net/api_tokens"
	if buildinfo.IsDevelopment() {
		// Extract host from apiBase for the token URL
		tokenURL = fmt.Sprintf("http://localhost:%d/api_tokens", getAPIPortFromBase(apiBase))
	}

	fmt.Println("Login to snet")
	fmt.Println("")
	fmt.Printf("Create an API token at: %s\n", tokenURL)
	fmt.Println("")
	fmt.Print("Enter your API token: ")

	token, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	token = strings.TrimSpace(token)

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Verify token by fetching accounts
	fmt.Println("")
	fmt.Printf("Verifying token against %s...\n", apiBase)

	client := api.NewClientWithToken(apiBase, token)
	accounts, err := client.ListAccounts()
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no accounts found for this token")
	}

	// Select account
	var selectedAccount api.Account
	if len(accounts) == 1 {
		selectedAccount = accounts[0]
		fmt.Printf("Using account: %s (%s)\n", selectedAccount.Name, selectedAccount.Slug)
	} else {
		fmt.Println("")
		fmt.Println("Select an account:")
		for i, acc := range accounts {
			fmt.Printf("  [%d] %s (%s)\n", i+1, acc.Name, acc.Slug)
		}
		fmt.Print("Enter number: ")

		selectionStr, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read selection: %w", err)
		}
		selection, err := strconv.Atoi(strings.TrimSpace(selectionStr))
		if err != nil || selection < 1 || selection > len(accounts) {
			return fmt.Errorf("invalid selection")
		}
		selectedAccount = accounts[selection-1]
	}

	// Save config - use prefix_id for API calls
	accountID := selectedAccount.PrefixID
	if accountID == "" {
		// Fallback to regular ID if prefix_id not available
		accountID = selectedAccount.ID.String()
	}

	wildcardDefault := true
	cfg := &config.Config{
		APIToken:        token,
		AccountID:       accountID,
		APIBase:         apiBase,
		DefaultProvider: "frp",            // Default to FRP for new logins
		DefaultWildcard: &wildcardDefault, // Default to wildcard enabled
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("")
	fmt.Println("Login successful!")
	fmt.Printf("Account: %s (%s)\n", selectedAccount.Name, selectedAccount.Slug)
	fmt.Printf("Default provider: %s\n", cfg.DefaultProvider)
	fmt.Printf("Default wildcard: %v\n", cfg.WildcardDefault())
	fmt.Println("")
	fmt.Println("You can now use 'snet start 3000' to create a tunnel.")
	fmt.Println("Wildcards are enabled by default. Use --no-wildcard to disable.")

	return nil
}

// getAPIPortFromBase extracts the port from an API base URL
func getAPIPortFromBase(apiBase string) int {
	// Default to 3001 for development
	var port int
	_, err := fmt.Sscanf(apiBase, "http://localhost:%d", &port)
	if err != nil {
		return 3001
	}
	return port
}
