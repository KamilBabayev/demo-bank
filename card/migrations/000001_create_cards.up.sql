-- Create cards table
CREATE TABLE IF NOT EXISTS cards (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    card_number VARCHAR(20) NOT NULL,  -- Masked format: **** **** **** 1234
    card_number_hash VARCHAR(64) NOT NULL,  -- SHA-256 hash for lookups
    card_type VARCHAR(20) NOT NULL,  -- 'debit', 'credit', 'virtual'
    cardholder_name VARCHAR(100) NOT NULL,
    expiration_month INTEGER NOT NULL,
    expiration_year INTEGER NOT NULL,
    cvv_hash VARCHAR(64),  -- SHA-256 hash, never stored plain
    pin_hash VARCHAR(64),  -- SHA-256 hash, never stored plain
    status VARCHAR(20) DEFAULT 'active',  -- 'active', 'blocked', 'expired', 'cancelled'
    daily_limit DECIMAL(15,2) DEFAULT 5000.00,
    monthly_limit DECIMAL(15,2) DEFAULT 50000.00,
    per_transaction_limit DECIMAL(15,2) DEFAULT 2000.00,
    daily_used DECIMAL(15,2) DEFAULT 0.00,
    monthly_used DECIMAL(15,2) DEFAULT 0.00,
    last_usage_date DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for account lookups
CREATE INDEX idx_cards_account_id ON cards(account_id);
-- Index for card number hash lookups
CREATE INDEX idx_cards_card_number_hash ON cards(card_number_hash);
-- Index for status filtering
CREATE INDEX idx_cards_status ON cards(status);
-- Index for card type filtering
CREATE INDEX idx_cards_card_type ON cards(card_type);
-- Index for expiration checking
CREATE INDEX idx_cards_expiration ON cards(expiration_year, expiration_month);

-- Create updated_at trigger function (if not exists)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_cards_updated_at BEFORE UPDATE ON cards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE cards IS 'Stores bank card information linked to accounts';
COMMENT ON COLUMN cards.card_number IS 'Masked card number (only last 4 digits visible)';
COMMENT ON COLUMN cards.card_number_hash IS 'SHA-256 hash of full card number for secure lookups';
COMMENT ON COLUMN cards.card_type IS 'Card type: debit, credit, or virtual';
COMMENT ON COLUMN cards.status IS 'Card status: active, blocked, expired, or cancelled';
COMMENT ON COLUMN cards.cvv_hash IS 'SHA-256 hash of CVV (never stored plain)';
COMMENT ON COLUMN cards.pin_hash IS 'SHA-256 hash of PIN (never stored plain)';
COMMENT ON COLUMN cards.daily_limit IS 'Maximum daily transaction limit';
COMMENT ON COLUMN cards.monthly_limit IS 'Maximum monthly transaction limit';
COMMENT ON COLUMN cards.per_transaction_limit IS 'Maximum per-transaction limit';
COMMENT ON COLUMN cards.daily_used IS 'Amount used today';
COMMENT ON COLUMN cards.monthly_used IS 'Amount used this month';
COMMENT ON COLUMN cards.last_usage_date IS 'Date of last card usage (for resetting limits)';
