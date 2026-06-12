-- Retention policy in days, stored as a single app_settings row.
SELECT setting_value
FROM app_settings
WHERE setting_key = 'trash_retention_days';
