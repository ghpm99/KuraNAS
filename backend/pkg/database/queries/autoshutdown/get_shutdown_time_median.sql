-- Median time-of-day (seconds since midnight, server timezone) of every recorded
-- SHUTDOWN, used to suggest an auto-shutdown time. percentile_cont interpolates,
-- so it is a true median. The time-of-day is not circular here: shutdowns that
-- straddle midnight skew the suggestion, which is acceptable for a hint.
SELECT
    percentile_cont(0.5) WITHIN GROUP (
        ORDER BY EXTRACT(EPOCH FROM event_time::time)
    ) AS median_seconds,
    COUNT(*) AS sample_size
FROM system_event_log
WHERE event_type = 'SHUTDOWN';
