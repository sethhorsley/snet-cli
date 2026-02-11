package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/seth4242/snet/internal/api"
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

	fmt.Println("Login to seth4242.net")
	fmt.Println("")
	fmt.Println("Create an API token at: https://seth4242.net/api_tokens")
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
	fmt.Println("Verifying token...")

	client := api.NewClientWithToken(config.DefaultAPIBase, token)
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

		var selection int
		_, err := fmt.Scanf("%d", &selection)
		if err != nil || selection < 1 || selection > len(accounts) {
			return fmt.Errorf("invalid selection")
		}
		selectedAccount = accounts[selection-1]
	}

	// Save config
	cfg := &config.Config{
		APIToken:  token,
		AccountID: selectedAccount.ID,
		APIBase:   config.DefaultAPIBase,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("")
	fmt.Println("Login successful!")
	fmt.Printf("Account: %s (%s)\n", selectedAccount.Name, selectedAccount.Slug)
	fmt.Println("")
	fmt.Println("You can now use 'snet start --port 3000' to create a tunnel.")

	return nil
}
