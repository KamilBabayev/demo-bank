package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage bank accounts",
	Long:  `List, view, deposit to, and withdraw from bank accounts.`,
}

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		resp, err := client.ListAccounts()
		if err != nil {
			return fmt.Errorf("failed to list accounts: %w", err)
		}

		if jsonOutput {
			printJSON(resp)
			return nil
		}

		if len(resp.Accounts) == 0 {
			fmt.Println("No accounts found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Account Number", "Type", "Balance", "Currency", "Status"})
		table.SetBorder(false)

		for _, a := range resp.Accounts {
			table.Append([]string{
				strconv.FormatInt(a.ID, 10),
				a.AccountNumber,
				a.AccountType,
				a.Balance,
				a.Currency,
				a.Status,
			})
		}

		table.Render()
		fmt.Printf("\nTotal: %d accounts\n", resp.Total)
		return nil
	},
}

var accountsViewCmd = &cobra.Command{
	Use:   "view <account_id>",
	Short: "View account details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid account ID: %w", err)
		}

		account, err := client.GetAccount(id)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		if jsonOutput {
			printJSON(account)
			return nil
		}

		fmt.Printf("Account ID:     %d\n", account.ID)
		fmt.Printf("Account Number: %s\n", account.AccountNumber)
		fmt.Printf("Type:           %s\n", account.AccountType)
		fmt.Printf("Balance:        %s %s\n", account.Balance, account.Currency)
		fmt.Printf("Status:         %s\n", account.Status)
		fmt.Printf("Created:        %s\n", account.CreatedAt)
		return nil
	},
}

var accountsBalanceCmd = &cobra.Command{
	Use:   "balance <account_id>",
	Short: "Show account balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid account ID: %w", err)
		}

		account, err := client.GetAccount(id)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		if jsonOutput {
			printJSON(map[string]interface{}{
				"account_id":     account.ID,
				"account_number": account.AccountNumber,
				"balance":        account.Balance,
				"currency":       account.Currency,
			})
			return nil
		}

		fmt.Printf("%s %s\n", account.Balance, account.Currency)
		return nil
	},
}

var depositAmount string

var accountsDepositCmd = &cobra.Command{
	Use:   "deposit <account_id>",
	Short: "Deposit money into an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		if depositAmount == "" {
			return fmt.Errorf("amount is required (--amount)")
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid account ID: %w", err)
		}

		account, err := client.Deposit(id, depositAmount)
		if err != nil {
			return fmt.Errorf("deposit failed: %w", err)
		}

		if jsonOutput {
			printJSON(map[string]interface{}{
				"message":     "Deposit successful",
				"new_balance": account.Balance,
				"currency":    account.Currency,
			})
			return nil
		}

		fmt.Printf("Deposit successful. New balance: %s %s\n", account.Balance, account.Currency)
		return nil
	},
}

var withdrawAmount string

var accountsWithdrawCmd = &cobra.Command{
	Use:   "withdraw <account_id>",
	Short: "Withdraw money from an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		if withdrawAmount == "" {
			return fmt.Errorf("amount is required (--amount)")
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid account ID: %w", err)
		}

		account, err := client.Withdraw(id, withdrawAmount)
		if err != nil {
			return fmt.Errorf("withdrawal failed: %w", err)
		}

		if jsonOutput {
			printJSON(map[string]interface{}{
				"message":     "Withdrawal successful",
				"new_balance": account.Balance,
				"currency":    account.Currency,
			})
			return nil
		}

		fmt.Printf("Withdrawal successful. New balance: %s %s\n", account.Balance, account.Currency)
		return nil
	},
}

func init() {
	accountsDepositCmd.Flags().StringVar(&depositAmount, "amount", "", "Amount to deposit")
	accountsWithdrawCmd.Flags().StringVar(&withdrawAmount, "amount", "", "Amount to withdraw")

	accountsCmd.AddCommand(accountsListCmd)
	accountsCmd.AddCommand(accountsViewCmd)
	accountsCmd.AddCommand(accountsBalanceCmd)
	accountsCmd.AddCommand(accountsDepositCmd)
	accountsCmd.AddCommand(accountsWithdrawCmd)

	rootCmd.AddCommand(accountsCmd)
}
