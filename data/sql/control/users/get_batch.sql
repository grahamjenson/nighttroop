--IN: batch string
--OUT: Subdomain string, LastSeenAt *time.Time
SELECT subdomain, last_seen_at FROM users WHERE batch = ? AND deleted_at IS NULL ORDER BY subdomain DESC