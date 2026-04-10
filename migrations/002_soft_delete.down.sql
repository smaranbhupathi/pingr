DROP INDEX IF EXISTS idx_monitors_not_deleted;
ALTER TABLE monitors DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_alert_channels_not_deleted;
ALTER TABLE alert_channels DROP COLUMN IF EXISTS deleted_at;
