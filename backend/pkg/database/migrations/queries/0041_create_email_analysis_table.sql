-- AI verdict for one synced message (task 16). The e-mail body is adversarial
-- input to the LLM, so the contract is fail-closed: an unparseable model answer
-- lands here as verdict='suspicious' with evidence=["ANALYSIS_PARSE_FAILED"].
-- evidence holds the structured signals behind the verdict (deterministic facts
-- + model rationale); summary is filled only for legitimate mail. provider_used
-- / model_used record which AI answered, for observability. There is at most one
-- analysis per message (UNIQUE), so a re-run upserts in place.
CREATE TABLE IF NOT EXISTS email_analysis (
    id SERIAL PRIMARY KEY,
    message_id INTEGER NOT NULL UNIQUE REFERENCES email_message (id) ON DELETE CASCADE,
    verdict VARCHAR(16) NOT NULL CHECK (verdict IN ('legitimate', 'suspicious', 'malicious')),
    risk_score INTEGER NOT NULL DEFAULT 0 CHECK (risk_score >= 0 AND risk_score <= 100),
    evidence JSONB NOT NULL DEFAULT '[]',
    summary TEXT NOT NULL DEFAULT '',
    importance VARCHAR(16) NOT NULL DEFAULT 'normal' CHECK (importance IN ('low', 'normal', 'high')),
    provider_used VARCHAR(32) NOT NULL DEFAULT '',
    model_used VARCHAR(128) NOT NULL DEFAULT '',
    analyzed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
