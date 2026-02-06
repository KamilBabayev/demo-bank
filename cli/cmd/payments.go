package cmd

import (
	"fmt"
	"os"
	"strconv"

	"dbank/api"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var paymentsCmd = &cobra.Command{
	Use:   "payments",
	Short: "Manage payments",
	Long:  `List and create payments.`,
}

var paymentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all payments",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		resp, err := client.ListPayments()
		if err != nil {
			return fmt.Errorf("failed to list payments: %w", err)
		}

		if jsonOutput {
			printJSON(resp)
			return nil
		}

		if len(resp.Payments) == 0 {
			fmt.Println("No payments found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Reference", "Account", "Type", "Recipient", "Amount", "Status"})
		table.SetBorder(false)

		for _, p := range resp.Payments {
			recipient := p.RecipientName
			if recipient == "" {
				recipient = "-"
			}
			ref := p.ReferenceID
			if len(ref) > 8 {
				ref = ref[:8] + "..."
			}
			table.Append([]string{
				strconv.FormatInt(p.ID, 10),
				ref,
				strconv.FormatInt(p.AccountID, 10),
				p.PaymentType,
				recipient,
				p.Amount + " " + p.Currency,
				p.Status,
			})
		}

		table.Render()
		fmt.Printf("\nTotal: %d payments\n", resp.Total)
		return nil
	},
}

var (
	paymentAccountID     int64
	paymentType          string
	paymentRecipient     string
	paymentRecipientAcct string
	paymentAmount        string
	paymentCurrency      string
	paymentDescription   string
)

var paymentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new payment",
	Long: `Create a new payment.

Payment types: bill, merchant, external

Examples:
  dbank payments create --account 1 --type bill --recipient "Electric Company" --amount 150
  dbank payments create --account 1 --type merchant --recipient "Amazon" --amount 50
  dbank payments create --account 1 --type external --recipient "John Doe" --recipient-account "123456789" --amount 200`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		if paymentAccountID == 0 {
			return fmt.Errorf("account ID is required (--account)")
		}
		if paymentType == "" {
			return fmt.Errorf("payment type is required (--type: bill, merchant, external)")
		}
		if paymentAmount == "" {
			return fmt.Errorf("amount is required (--amount)")
		}

		req := &api.CreatePaymentRequest{
			AccountID:        paymentAccountID,
			PaymentType:      paymentType,
			RecipientName:    paymentRecipient,
			RecipientAccount: paymentRecipientAcct,
			Amount:           paymentAmount,
			Currency:         paymentCurrency,
			Description:      paymentDescription,
		}

		payment, err := client.CreatePayment(req)
		if err != nil {
			return fmt.Errorf("payment failed: %w", err)
		}

		if jsonOutput {
			printJSON(payment)
			return nil
		}

		fmt.Printf("Payment created successfully\n")
		fmt.Printf("Payment ID:   %d\n", payment.ID)
		fmt.Printf("Reference:    %s\n", payment.ReferenceID)
		fmt.Printf("Status:       %s\n", payment.Status)
		return nil
	},
}

func init() {
	paymentsCreateCmd.Flags().Int64Var(&paymentAccountID, "account", 0, "Source account ID")
	paymentsCreateCmd.Flags().StringVar(&paymentType, "type", "", "Payment type (bill, merchant, external)")
	paymentsCreateCmd.Flags().StringVar(&paymentRecipient, "recipient", "", "Recipient name")
	paymentsCreateCmd.Flags().StringVar(&paymentRecipientAcct, "recipient-account", "", "Recipient account number (for external transfers)")
	paymentsCreateCmd.Flags().StringVar(&paymentAmount, "amount", "", "Payment amount")
	paymentsCreateCmd.Flags().StringVar(&paymentCurrency, "currency", "USD", "Currency (default: USD)")
	paymentsCreateCmd.Flags().StringVar(&paymentDescription, "description", "", "Payment description")

	paymentsCmd.AddCommand(paymentsListCmd)
	paymentsCmd.AddCommand(paymentsCreateCmd)

	rootCmd.AddCommand(paymentsCmd)
}
