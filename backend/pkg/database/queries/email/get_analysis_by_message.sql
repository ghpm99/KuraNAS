-- One message's AI verdict for the summary endpoint.
SELECT message_id, verdict, risk_score, evidence, summary, importance,
       provider_used, model_used
FROM email_analysis
WHERE message_id = $1;
