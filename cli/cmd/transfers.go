package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var transfersCmd = &cobra.Command{
	Use:   "transfers",
	Short: "Manage transfers",
	Long:  `List and create bank transfers.`,
}

var transfersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all transfers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		resp, err := client.ListTransfers()
		if err != nil {
			return fmt.Errorf("failed to list transfers: %w", err)
		}

		if jsonOutput {
			printJSON(resp)
			return nil
		}

		if len(resp.Transfers) == 0 {
			fmt.Println("No transfers found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Reference", "From", "To", "Amount", "Status", "Created"})
		table.SetBorder(false)

		for _, t := range resp.Transfers {
			ref := t.ReferenceID
			if len(ref) > 8 {
				ref = ref[:8] + "..."
			}
			created := t.CreatedAt
			if len(created) > 10 {
				created = created[:10]
			}
			table.Append([]string{
				strconv.FormatInt(t.ID, 10),
				ref,
				strconv.FormatInt(t.FromAccountID, 10),
				strconv.FormatInt(t.ToAccountID, 10),
				t.Amount + " " + t.Currency,
				t.Status,
				created,
			})
		}

		table.Render()
		fmt.Printf("\nTotal: %d transfers\n", resp.Total)
		return nil
	},
}

var transfersViewCmd = &cobra.Command{
	Use:   "view <transfer_id>",
	Short: "View transfer details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid transfer ID: %w", err)
		}

		transfer, err := client.GetTransfer(id)
		if err != nil {
			return fmt.Errorf("failed to get transfer: %w", err)
		}

		if jsonOutput {
			printJSON(transfer)
			return nil
		}

		fmt.Printf("Transfer ID:      %d\n", transfer.ID)
		fmt.Printf("Reference:        %s\n", transfer.ReferenceID)
		fmt.Printf("From Account:     %d\n", transfer.FromAccountID)
		fmt.Printf("To Account:       %d\n", transfer.ToAccountID)
		fmt.Printf("Amount:           %s %s\n", transfer.Amount, transfer.Currency)
		fmt.Printf("Status:           %s\n", transfer.Status)
		if transfer.FailureReason != "" {
			fmt.Printf("Failure Reason:   %s\n", transfer.FailureReason)
		}
		fmt.Printf("Created:          %s\n", transfer.CreatedAt)
		if transfer.CompletedAt != "" {
			fmt.Printf("Completed:        %s\n", transfer.CompletedAt)
		}
		return nil
	},
}

var (
	transferFrom     int64
	transferTo       int64
	transferAmount   string
	transferCurrency string
)

var transfersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new transfer",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		if transferFrom == 0 {
			return fmt.Errorf("source account is required (--from)")
		}
		if transferTo == 0 {
			return fmt.Errorf("destination account is required (--to)")
		}
		if transferAmount == "" {
			return fmt.Errorf("amount is required (--amount)")
		}

		resp, err := client.CreateTransfer(transferFrom, transferTo, transferAmount, transferCurrency)
		if err != nil {
			return fmt.Errorf("transfer failed: %w", err)
		}

		if jsonOutput {
			printJSON(resp)
			return nil
		}

		fmt.Printf("Transfer initiated successfully\n")
		fmt.Printf("Transfer ID:  %d\n", resp.TransferID)
		fmt.Printf("Reference:    %s\n", resp.ReferenceID)
		fmt.Printf("Status:       %s\n", resp.Status)
		return nil
	},
}

func init() {
	transfersCreateCmd.Flags().Int64Var(&transferFrom, "from", 0, "Source account ID")
	transfersCreateCmd.Flags().Int64Var(&transferTo, "to", 0, "Destination account ID")
	transfersCreateCmd.Flags().StringVar(&transferAmount, "amount", "", "Amount to transfer")
	transfersCreateCmd.Flags().StringVar(&transferCurrency, "currency", "USD", "Currency (default: USD)")

	transfersCmd.AddCommand(transfersListCmd)
	transfersCmd.AddCommand(transfersViewCmd)
	transfersCmd.AddCommand(transfersCreateCmd)

	rootCmd.AddCommand(transfersCmd)
}
