-- finds users where the last_seen_at is more than minutes ago and deleted_at is null
SELECT subdomain
FROM users 
WHERE deleted_at IS NULL AND last_seen_at < datetime('now', ? || ' minutes');
