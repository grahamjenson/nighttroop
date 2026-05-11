--IN: 
--OUT: Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time, DeletedAt *time.Time
SELECT subdomain, created_at, last_seen_at, deleted_at
FROM users WHERE deleted_at IS NULL;
