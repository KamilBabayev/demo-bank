-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,  -- 'transfer_sent', 'transfer_received', 'payment_processed', etc.
    channel VARCHAR(20) NOT NULL,  -- 'email', 'sms', 'push'
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB,  -- Additional data like transfer_id, payment_id, etc.
    status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'sent', 'failed', 'read'
    read_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_user_status ON notifications(user_id, status);

-- Create updated_at trigger function (if not exists)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE notifications IS 'Stores user notifications from various events';
COMMENT ON COLUMN notifications.type IS 'Type of notification: transfer_sent, transfer_received, payment_processed, etc.';
COMMENT ON COLUMN notifications.channel IS 'Delivery channel: email, sms, or push';
COMMENT ON COLUMN notifications.status IS 'Notification status: pending, sent, failed, or read';
COMMENT ON COLUMN notifications.metadata IS 'JSON metadata with event-specific information';
