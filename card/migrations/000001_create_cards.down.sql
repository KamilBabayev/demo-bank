-- Drop trigger first
DROP TRIGGER IF EXISTS update_cards_updated_at ON cards;

-- Drop indexes
DROP INDEX IF EXISTS idx_cards_account_id;
DROP INDEX IF EXISTS idx_cards_card_number_hash;
DROP INDEX IF EXISTS idx_cards_status;
DROP INDEX IF EXISTS idx_cards_card_type;
DROP INDEX IF EXISTS idx_cards_expiration;

-- Drop table
DROP TABLE IF EXISTS cards;
