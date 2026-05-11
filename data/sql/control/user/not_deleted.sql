--IN: 
--OUT: subdomain string, created_at *time.Time, last_seen_at *time.Time, deleted_at *time.Time
SELECT subdomain, created_at, last_seen_at, deleted_at
FROM users WHERE deleted_at IS NULL;
