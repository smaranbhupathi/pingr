-- Soft delete for monitors
ALTER TABLE monitors ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_monitors_not_deleted ON monitors(user_id) WHERE deleted_at IS NULL;

-- Soft delete for alert_channels
ALTER TABLE alert_channels ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_alert_channels_not_deleted ON alert_channels(user_id) WHERE deleted_at IS NULL;
