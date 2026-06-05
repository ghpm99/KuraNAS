-- The original constraint (migration 0022) only allowed STARTUP/SHUTDOWN, but the
-- observability overhaul added more event types in pkg/systemevent/model.go. Every
-- INSERT of a new type was being rejected, so the events never reached the log.
ALTER TABLE system_event_log
    DROP CONSTRAINT IF EXISTS system_event_log_event_type_check;

ALTER TABLE system_event_log
    ADD CONSTRAINT system_event_log_event_type_check
    CHECK (event_type IN (
        'STARTUP',
        'SHUTDOWN',
        'WORKER_POOL_STARTED',
        'SCAN_COMPLETED',
        'JOB_FAILED',
        'AI_PROVIDER_UNAVAILABLE'
    ));
