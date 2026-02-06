-- Drop trigger first
DROP TRIGGER IF EXISTS update_transfers_updated_at ON transfers;

-- Drop indexes
DROP INDEX IF EXISTS idx_transfers_reference_id;
DROP INDEX IF EXISTS idx_transfers_from_account_id;
DROP INDEX IF EXISTS idx_transfers_to_account_id;
DROP INDEX IF EXISTS idx_transfers_status;
DROP INDEX IF EXISTS idx_transfers_created_at;

-- Drop table
DROP TABLE IF EXISTS transfers;
