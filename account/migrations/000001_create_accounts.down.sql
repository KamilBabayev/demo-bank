-- Drop trigger first
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;

-- Drop indexes
DROP INDEX IF EXISTS idx_accounts_user_id;
DROP INDEX IF EXISTS idx_accounts_account_number;
DROP INDEX IF EXISTS idx_accounts_status;
DROP INDEX IF EXISTS idx_accounts_account_type;

-- Drop table
DROP TABLE IF EXISTS accounts;
