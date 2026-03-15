INSERT INTO app_settings (
    setting_key,
    setting_value,
    created_at,
    updated_at
)
VALUES (
    $1,
    $2,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (setting_key) DO UPDATE
SET
    setting_value = EXCLUDED.setting_value,
    updated_at = CURRENT_TIMESTAMP;
