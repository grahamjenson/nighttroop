-- finds users where the last_seen_at is more than 30 days ago and deleted_at is null
SELECT subdomain
FROM users
WHERE last_seen_at < datetime('now', '-30 days')
    AND deleted_at IS NULL;