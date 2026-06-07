-- The Ollama daemon autostart (boot lifecycle) records two new event types.
-- Recreate the check constraint so these INSERTs are accepted, otherwise the
-- events would be silently rejected like the ones migration 0032 fixed.
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
        'AI_PROVIDER_UNAVAILABLE',
        'OLLAMA_DAEMON_STARTED',
        'OLLAMA_DAEMON_UNREACHABLE'
    ));
