package cmd

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	username string
	password string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Demo Bank",
	Long:  `Authenticate with your Demo Bank credentials and store the session token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if username == "" {
			return fmt.Errorf("username is required (-u or --username)")
		}

		// If password not provided via flag, prompt for it
		if password == "" {
			fmt.Print("Password: ")
			pwBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println()
			password = string(pwBytes)
		}

		resp, err := client.Login(username, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Save token and user info to config
		cfg.Token = resp.Token
		cfg.Username = resp.User.Username
		cfg.UserID = resp.User.ID
		cfg.Role = resp.User.Role

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if jsonOutput {
			printJSON(map[string]interface{}{
				"message":  "Login successful",
				"username": resp.User.Username,
				"role":     resp.User.Role,
			})
		} else {
			fmt.Printf("Logged in as %s (%s)\n", resp.User.Username, resp.User.Role)
		}

		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Demo Bank",
	Long:  `Clear the stored session token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfg.Clear(); err != nil {
			return fmt.Errorf("failed to clear config: %w", err)
		}

		if jsonOutput {
			printJSON(map[string]string{"message": "Logged out successfully"})
		} else {
			fmt.Println("Logged out successfully")
		}

		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current logged in user",
	Long:  `Display information about the currently authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cfg.IsLoggedIn() {
			if jsonOutput {
				printJSON(map[string]string{"message": "Not logged in"})
			} else {
				fmt.Println("Not logged in")
			}
			return nil
		}

		if jsonOutput {
			printJSON(map[string]interface{}{
				"username": cfg.Username,
				"user_id":  cfg.UserID,
				"role":     cfg.Role,
				"api_url":  cfg.APIURL,
			})
		} else {
			fmt.Printf("Username: %s\n", cfg.Username)
			fmt.Printf("User ID:  %d\n", cfg.UserID)
			fmt.Printf("Role:     %s\n", cfg.Role)
			fmt.Printf("API URL:  %s\n", cfg.APIURL)
		}

		return nil
	},
}

func init() {
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password (will prompt if not provided)")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
}
