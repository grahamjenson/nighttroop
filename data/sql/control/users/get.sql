
--IN: Subdomain string
--OUT: Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time

SELECT subdomain, created_at, last_seen_at
FROM users
WHERE subdomain = ? AND deleted_at IS NULL;
