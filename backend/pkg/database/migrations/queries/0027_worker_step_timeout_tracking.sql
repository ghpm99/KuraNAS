-- Package 1: step timeout + back-of-the-line requeue + recurring-timeout visibility.
--
-- timeout_count: how many times a step was deferred to the back of the queue
-- because it exceeded the execution timeout. It never causes a hard failure
-- (infinite circulation), so this counter is the signal to watch: a value that
-- keeps climbing means a file is too slow or corrupted and needs attention.
ALTER TABLE worker_step
    ADD COLUMN IF NOT EXISTS timeout_count INTEGER NOT NULL DEFAULT 0;

-- next_attempt_at: when set, the job re-enters the queue ordered by this time
-- instead of created_at, sending a deferred job to the *back* of the line
-- (FIFO with re-queue at the tail).
ALTER TABLE worker_job
    ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMPTZ;

-- Supports the scheduler's FIFO ordering: COALESCE(next_attempt_at, created_at).
CREATE INDEX IF NOT EXISTS idx_worker_job_status_next_attempt
    ON worker_job (status, (COALESCE(next_attempt_at, created_at)));

-- Fast lookup of files stuck in recurring timeouts for the analytics card.
CREATE INDEX IF NOT EXISTS idx_worker_step_timeout_count
    ON worker_step (timeout_count)
    WHERE timeout_count > 0;
