-- Sets the retention policy in days.
INSERT INTO app_settings (setting_key, setting_value)
VALUES ('trash_retention_days', $1)
ON CONFLICT (setting_key)
DO UPDATE SET setting_value = EXCLUDED.setting_value, updated_at = CURRENT_TIMESTAMP;
