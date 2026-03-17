DELETE FROM notifications
WHERE created_at < NOW() - INTERVAL '30 days';
