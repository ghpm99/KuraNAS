INSERT INTO system_event_log (
    event_time,
    event_time_display,
    event_type,
    description,
    source,
    host_name,
    process_id,
    extra_data
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
