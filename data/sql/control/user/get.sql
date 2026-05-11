
--IN: subdomain string
--OUT: subdomain string, created_at *time.Time, last_seen_at *time.Time

SELECT subdomain, created_at, last_seen_at
FROM users
WHERE subdomain = ? AND deleted_at IS NULL;
