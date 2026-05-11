--IN: batch string
--OUT: subdomain string, last_seen_at *time.Time
SELECT subdomain, last_seen_at FROM users WHERE batch = ? AND deleted_at IS NULL ORDER BY subdomain DESC