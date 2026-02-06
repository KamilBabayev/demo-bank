package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"dbank/api"
	"dbank/config"

	"github.com/spf13/cobra"
)

var (
	cfg        *config.Config
	client     *api.Client
	jsonOutput bool
)

var rootCmd = &cobra.Command{
	Use:   "dbank",
	Short: "CLI client for Demo Bank",
	Long: `A command-line interface for Demo Bank operations.

Use this tool to manage accounts, transfers, payments, and notifications
from your terminal.

Get started by logging in:
  dbank login -u <username> -p <password>

Then run commands like:
  dbank accounts list
  dbank transfers create --from 1 --to 2 --amount 100
  dbank notifications list`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if API URL override is set
		if apiURL := os.Getenv("DBANK_API_URL"); apiURL != "" {
			cfg.APIURL = apiURL
		}

		client = api.NewClient(cfg)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
}

// Helper functions for output

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func requireAuth() error {
	if !cfg.IsLoggedIn() {
		return fmt.Errorf("not logged in. Run: dbank login -u <username> -p <password>")
	}
	return nil
}
