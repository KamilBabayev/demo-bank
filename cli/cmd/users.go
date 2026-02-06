package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users (admin only)",
	Long:  `List and view users. Requires admin role.`,
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		resp, err := client.ListUsers()
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		if jsonOutput {
			printJSON(resp)
			return nil
		}

		if len(resp.Users) == 0 {
			fmt.Println("No users found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Username", "Email", "Name", "Role", "Status"})
		table.SetBorder(false)

		for _, u := range resp.Users {
			name := u.FirstName
			if u.LastName != "" {
				name += " " + u.LastName
			}
			table.Append([]string{
				strconv.FormatInt(u.ID, 10),
				u.Username,
				u.Email,
				name,
				u.Role,
				u.Status,
			})
		}

		table.Render()
		fmt.Printf("\nTotal: %d users\n", resp.Total)
		return nil
	},
}

func init() {
	usersCmd.AddCommand(usersListCmd)
	rootCmd.AddCommand(usersCmd)
}
