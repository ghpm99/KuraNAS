-- Stores one message's AI verdict. UNIQUE(message_id) makes a re-analysis
-- replace the previous verdict in place (idempotent retries).
INSERT INTO email_analysis (
    message_id, verdict, risk_score, evidence, summary, importance,
    provider_used, model_used, analyzed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
ON CONFLICT (message_id) DO UPDATE
SET verdict = EXCLUDED.verdict,
    risk_score = EXCLUDED.risk_score,
    evidence = EXCLUDED.evidence,
    summary = EXCLUDED.summary,
    importance = EXCLUDED.importance,
    provider_used = EXCLUDED.provider_used,
    model_used = EXCLUDED.model_used,
    analyzed_at = now();
