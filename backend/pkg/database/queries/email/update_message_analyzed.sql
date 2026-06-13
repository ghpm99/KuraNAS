-- Post-analysis retention (hard rule A7): once a message has a verdict the
-- sanitized body is no longer needed and is expunged — only snippet + summary
-- remain. Also advances the status (analyzed | failed).
UPDATE email_message
SET status = $2, sanitized_body = NULL
WHERE id = $1;
