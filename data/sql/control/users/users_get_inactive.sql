-- Inactive Users are users that were once active but have stopped, so the assumption is they have data 
SELECT subdomain
FROM users 
WHERE deleted_at IS NULL AND (last_seen_at < datetime('now', '-' || ? || ' minutes') OR last_write_at < datetime('now', '-' || ? || ' minutes'));