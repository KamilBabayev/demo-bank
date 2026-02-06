package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var notificationsCmd = &cobra.Command{
	Use:     "notifications",
	Aliases: []string{"notif"},
	Short:   "Manage notifications",
	Long:    `List and manage notifications.`,
}

var notificationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		resp, err := client.ListNotifications()
		if err != nil {
			return fmt.Errorf("failed to list notifications: %w", err)
		}

		if jsonOutput {
			printJSON(resp)
			return nil
		}

		if len(resp.Notifications) == 0 {
			fmt.Println("No notifications found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Type", "Title", "Status", "Created"})
		table.SetBorder(false)

		for _, n := range resp.Notifications {
			status := n.Status
			if status == "pending" {
				status = "unread"
			}
			created := n.CreatedAt
			if len(created) > 10 {
				created = created[:10]
			}
			table.Append([]string{
				strconv.FormatInt(n.ID, 10),
				n.Type,
				truncate(n.Title, 30),
				status,
				created,
			})
		}

		table.Render()
		fmt.Printf("\nTotal: %d (%d unread)\n", resp.Total, resp.Unread)
		return nil
	},
}

var notificationsReadCmd = &cobra.Command{
	Use:   "read <notification_id>",
	Short: "Mark a notification as read",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid notification ID: %w", err)
		}

		if err := client.MarkNotificationRead(id); err != nil {
			return fmt.Errorf("failed to mark notification as read: %w", err)
		}

		if jsonOutput {
			printJSON(map[string]string{"message": "Notification marked as read"})
		} else {
			fmt.Println("Notification marked as read")
		}
		return nil
	},
}

var notificationsReadAllCmd = &cobra.Command{
	Use:   "read-all",
	Short: "Mark all notifications as read",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		if err := client.MarkAllNotificationsRead(); err != nil {
			return fmt.Errorf("failed to mark notifications as read: %w", err)
		}

		if jsonOutput {
			printJSON(map[string]string{"message": "All notifications marked as read"})
		} else {
			fmt.Println("All notifications marked as read")
		}
		return nil
	},
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	notificationsCmd.AddCommand(notificationsListCmd)
	notificationsCmd.AddCommand(notificationsReadCmd)
	notificationsCmd.AddCommand(notificationsReadAllCmd)

	rootCmd.AddCommand(notificationsCmd)
}
