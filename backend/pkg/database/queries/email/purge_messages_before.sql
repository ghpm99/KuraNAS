-- Retention expurgo: drop messages older than the cutoff.
DELETE FROM email_message WHERE received_at < $1;
