SELECT
    setting_key,
    setting_value
FROM
    app_settings
WHERE
    setting_key = $1;
